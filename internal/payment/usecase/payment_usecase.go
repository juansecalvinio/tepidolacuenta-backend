package usecase

import (
	"context"
	"errors"
	"fmt"

	mpClient "juansecalvinio/tepidolacuenta/internal/payment/infrastructure/mercadopago"
	"juansecalvinio/tepidolacuenta/internal/payment/domain"
	"juansecalvinio/tepidolacuenta/internal/payment/repository"
	"juansecalvinio/tepidolacuenta/internal/pkg"
	restaurantRepo "juansecalvinio/tepidolacuenta/internal/restaurant/repository"
	subscriptionDomain "juansecalvinio/tepidolacuenta/internal/subscription/domain"
	subscriptionRepo "juansecalvinio/tepidolacuenta/internal/subscription/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UseCase interface {
	CreatePaymentPreference(ctx context.Context, userID primitive.ObjectID, input domain.CreatePreferenceInput) (*CreatePreferenceResult, error)
	ProcessPaymentWebhook(ctx context.Context, paymentID, xSignature, xRequestID string) error
	GetPaymentByID(ctx context.Context, userID primitive.ObjectID, paymentID string) (*domain.Payment, error)
	GetPaymentHistory(ctx context.Context, userID primitive.ObjectID, restaurantID string) ([]*domain.Payment, error)
}

type CreatePreferenceResult struct {
	PaymentURL   string
	PreferenceID string
}

// NotifyFunc is called after a payment is approved to notify connected clients via WebSocket
type NotifyFunc func(restaurantID primitive.ObjectID, event *domain.PaymentEvent)

type paymentUseCase struct {
	paymentRepo      repository.Repository
	planRepo         subscriptionRepo.PlanRepository
	subscriptionRepo subscriptionRepo.SubscriptionRepository
	restaurantRepo   restaurantRepo.Repository
	mp               *mpClient.Client
	notificationURL  string
	frontendURL      string
	notify           NotifyFunc
}

func NewPaymentUseCase(
	paymentRepo repository.Repository,
	planRepo subscriptionRepo.PlanRepository,
	subscriptionRepo subscriptionRepo.SubscriptionRepository,
	restaurantRepo restaurantRepo.Repository,
	mp *mpClient.Client,
	notificationURL string,
	frontendURL string,
	notify NotifyFunc,
) UseCase {
	return &paymentUseCase{
		paymentRepo:      paymentRepo,
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
		restaurantRepo:   restaurantRepo,
		mp:               mp,
		notificationURL:  notificationURL,
		frontendURL:      frontendURL,
		notify:           notify,
	}
}

// CreatePaymentPreference validates ownership, creates a MercadoPago preference and saves a pending Payment
func (uc *paymentUseCase) CreatePaymentPreference(ctx context.Context, userID primitive.ObjectID, input domain.CreatePreferenceInput) (*CreatePreferenceResult, error) {
	restaurantID, err := primitive.ObjectIDFromHex(input.RestaurantID)
	if err != nil {
		return nil, errors.New("invalid restaurant ID")
	}

	planID, err := primitive.ObjectIDFromHex(input.PlanID)
	if err != nil {
		return nil, errors.New("invalid plan ID")
	}

	restaurant, err := uc.restaurantRepo.FindByID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	plan, err := uc.planRepo.FindByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	// Save pending payment first to get the MongoDB ID for external_reference
	payment := domain.NewPayment(userID, restaurantID, planID, "", plan.Price)
	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	prefResp, err := uc.mp.CreatePreference(ctx, mpClient.PreferenceRequest{
		Title:             fmt.Sprintf("Plan %s - TePidoLaCuenta", plan.Name),
		Quantity:          1,
		UnitPrice:         plan.Price,
		ExternalReference: payment.ID.Hex(),
		NotificationURL:   uc.notificationURL,
		BackURLSuccess:    uc.frontendURL + "/payment/success",
		BackURLFailure:    uc.frontendURL + "/payment/failure",
		BackURLPending:    uc.frontendURL + "/payment/pending",
	})
	if err != nil {
		return nil, err
	}

	// Update payment with the preference ID
	payment.MPPreferenceID = prefResp.ID
	if err := uc.paymentRepo.Update(ctx, payment); err != nil {
		return nil, err
	}

	return &CreatePreferenceResult{
		PaymentURL:   prefResp.InitPoint,
		PreferenceID: prefResp.ID,
	}, nil
}

// ProcessPaymentWebhook handles the MercadoPago webhook notification
func (uc *paymentUseCase) ProcessPaymentWebhook(ctx context.Context, paymentID, xSignature, xRequestID string) error {
	log := pkg.NewLogger(ctx)

	if !uc.mp.ValidateWebhookSignature(xSignature, xRequestID, paymentID) {
		log.Warn("webhook signature validation failed for payment %s", paymentID)
		return pkg.ErrUnauthorized
	}

	mpPaymentID, err := parsePaymentID(paymentID)
	if err != nil {
		log.Error("invalid payment ID from webhook: %s", paymentID)
		return errors.New("invalid payment ID from webhook")
	}

	mpPayment, err := uc.mp.GetPayment(ctx, mpPaymentID)
	if err != nil {
		log.Error("failed to get payment %d from MercadoPago: %v", mpPaymentID, err)
		return err
	}

	if mpPayment.ExternalReference == "" {
		log.Warn("webhook received for MP payment %d with empty external_reference, skipping", mpPaymentID)
		return nil
	}

	internalID, err := primitive.ObjectIDFromHex(mpPayment.ExternalReference)
	if err != nil {
		log.Error("invalid external_reference %q for MP payment %d", mpPayment.ExternalReference, mpPaymentID)
		return errors.New("invalid external reference")
	}

	payment, err := uc.paymentRepo.FindByID(ctx, internalID)
	if err != nil {
		log.Error("payment %s not found in DB for MP payment %d: %v", internalID.Hex(), mpPaymentID, err)
		return err
	}

	payment.MPPaymentID = fmt.Sprintf("%d", mpPayment.ID)

	log.Info("processing webhook for payment %s: MP status=%s", payment.ID.Hex(), mpPayment.Status)

	switch mpPayment.Status {
	case "approved":
		payment.Status = domain.PaymentStatusApproved
		if err := uc.paymentRepo.Update(ctx, payment); err != nil {
			log.Error("failed to update payment %s to approved: %v", payment.ID.Hex(), err)
			return err
		}
		if err := uc.activateSubscription(ctx, payment); err != nil {
			log.Error("failed to activate subscription for payment %s: %v", payment.ID.Hex(), err)
			return err
		}
		if uc.notify != nil {
			uc.notify(payment.RestaurantID, domain.NewPaymentApprovedEvent(payment))
		}
		log.Info("payment %s approved and subscription activated", payment.ID.Hex())
		return nil

	case "rejected":
		payment.Status = domain.PaymentStatusRejected
		if err := uc.paymentRepo.Update(ctx, payment); err != nil {
			log.Error("failed to update payment %s to rejected: %v", payment.ID.Hex(), err)
			return err
		}
		log.Info("payment %s marked as rejected", payment.ID.Hex())
		return nil

	case "cancelled":
		payment.Status = domain.PaymentStatusCancelled
		if err := uc.paymentRepo.Update(ctx, payment); err != nil {
			log.Error("failed to update payment %s to cancelled: %v", payment.ID.Hex(), err)
			return err
		}
		log.Info("payment %s marked as cancelled", payment.ID.Hex())
		return nil

	default:
		// pending or other transient states — no action needed
		log.Info("payment %s has transient status %q, no action taken", payment.ID.Hex(), mpPayment.Status)
		return nil
	}
}

// GetPaymentByID returns a single payment by its ID, validating user ownership via the associated restaurant
func (uc *paymentUseCase) GetPaymentByID(ctx context.Context, userID primitive.ObjectID, paymentIDStr string) (*domain.Payment, error) {
	paymentID, err := primitive.ObjectIDFromHex(paymentIDStr)
	if err != nil {
		return nil, errors.New("invalid payment ID")
	}

	payment, err := uc.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	restaurant, err := uc.restaurantRepo.FindByID(ctx, payment.RestaurantID)
	if err != nil {
		return nil, err
	}

	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	return payment, nil
}

// GetPaymentHistory returns all payments for a restaurant, validating ownership
func (uc *paymentUseCase) GetPaymentHistory(ctx context.Context, userID primitive.ObjectID, restaurantIDStr string) ([]*domain.Payment, error) {
	restaurantID, err := primitive.ObjectIDFromHex(restaurantIDStr)
	if err != nil {
		return nil, errors.New("invalid restaurant ID")
	}

	restaurant, err := uc.restaurantRepo.FindByID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	return uc.paymentRepo.FindByRestaurantID(ctx, restaurantID)
}

// activateSubscription creates a new active subscription after a successful payment
func (uc *paymentUseCase) activateSubscription(ctx context.Context, payment *domain.Payment) error {
	// Check if an active subscription already exists (idempotency)
	existing, err := uc.subscriptionRepo.FindByRestaurantID(ctx, payment.RestaurantID)
	if err == nil && existing != nil &&
		existing.Status != subscriptionDomain.SubscriptionStatusCanceled &&
		existing.Status != subscriptionDomain.SubscriptionStatusExpired {
		// Already active — update plan and payment ref
		existing.PlanID = payment.PlanID
		existing.Status = subscriptionDomain.SubscriptionStatusActive
		existing.PaymentSubscriptionID = payment.MPPaymentID
		return uc.subscriptionRepo.Update(ctx, existing)
	}

	subscription := subscriptionDomain.NewSubscription(
		payment.UserID,
		payment.RestaurantID,
		payment.PlanID,
		subscriptionDomain.SubscriptionStatusActive,
	)
	subscription.PaymentSubscriptionID = payment.MPPaymentID

	return uc.subscriptionRepo.Create(ctx, subscription)
}

func parsePaymentID(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}
