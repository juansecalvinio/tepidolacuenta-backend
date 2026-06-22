# Ambiente previo (staging)

Guía para levantar y mantener el ambiente de pruebas, aislado de producción.

## Arquitectura

| Recurso | Producción | Staging (previa) |
|---|---|---|
| Base de datos | `tepidolacuenta` | `tepidolacuenta_dev` (mismo cluster) |
| Contenedor backend | `tepidolacuenta-backend` :8081 | `tepidolacuenta-backend-staging` :8082 |
| Imagen Docker | `…/tepidolacuenta-backend:latest` | `…/tepidolacuenta-backend:staging` |
| Dominio API | `api.tepidolacuenta.site` | `api-dev.tepidolacuenta.site` |
| Frontend | `app.tepidolacuenta.site` (Vercel prod) | preview de Vercel (rama `staging`) |
| Deploy | push a `main` | push a `staging` |
| Sentry environment | `production` | `staging` |
| MercadoPago | credenciales productivas | credenciales de **test** |

El selector de base lo da la env `MONGODB_DATABASE` (default `tepidolacuenta`).
Ambos contenedores comparten el mismo cluster Mongo y el mismo EC2; lo único
que cambia es la base, el puerto, el tag de imagen y un puñado de secrets.

---

## 1. Secrets de GitHub (Settings → Secrets and variables → Actions)

El workflow `deploy-staging.yml` **reutiliza** estos secrets ya existentes:
`DOCKER_USERNAME`, `DOCKER_PASSWORD`, `EC2_HOST`, `EC2_SSH_KEY`, `JWT_SECRET`,
`MONGODB_URI`, `SMTP_*`, `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `SENTRY_DSN`.

Y necesita estos **nuevos** secrets propios de staging:

| Secret | Valor sugerido |
|---|---|
| `STAGING_FRONTEND_BASE_URL` | URL del front de previa (ver Vercel, paso 5). Ej: `https://tepidolacuenta-frontend-git-staging-<user>.vercel.app` |
| `STAGING_CORS_ALLOWED_ORIGINS` | mismo origin del front de previa (coma-separado si hay varios) |
| `STAGING_GOOGLE_REDIRECT_URL` | `https://api-dev.tepidolacuenta.site/api/v1/auth/google/callback` |
| `STAGING_MERCADOPAGO_ACCESS_TOKEN` | access token de **test** de tu app de MercadoPago |
| `STAGING_MERCADOPAGO_NOTIFICATION_URL` | `https://api-dev.tepidolacuenta.site/api/v1/payments/webhook` |
| `STAGING_MERCADOPAGO_WEBHOOK_SECRET` | secret del webhook de test |

> `MONGODB_DATABASE=tepidolacuenta_dev` y `SENTRY_ENVIRONMENT=staging` ya van
> hardcodeados en el workflow, no hacen falta como secrets.

---

## 2. DNS (Hostinger)

Crear un registro A para el subdominio de la API de previa, apuntando a la
**misma IP del EC2** que `api`:

```
Tipo: A   Nombre: api-dev   Valor: <IP pública del EC2>   TTL: 3600
```

---

## 3. nginx en el EC2 (reverse proxy + WebSocket)

Crear `/etc/nginx/sites-available/api-dev.tepidolacuenta.site` apuntando al
contenedor de staging (puerto 8082). Importante el bloque `Upgrade` para que
funcione el WebSocket de notificaciones:

```nginx
server {
    server_name api-dev.tepidolacuenta.site;

    location / {
        proxy_pass http://127.0.0.1:8082;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 3600s;
    }
}
```

Activar y recargar:

```bash
sudo ln -s /etc/nginx/sites-available/api-dev.tepidolacuenta.site /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

> El security group del EC2 no necesita cambios: nginx (80/443) hace de proxy,
> el puerto 8082 queda solo en localhost.

---

## 4. SSL (certbot)

```bash
sudo certbot --nginx -d api-dev.tepidolacuenta.site
```

---

## 5. Frontend en Vercel (previa)

El front en Vercel ya genera **preview deployments** por rama. Para que la rama
`staging` apunte al backend de previa, configurar variables de entorno con
scope **Preview** (o por rama `staging`):

| Variable | Valor |
|---|---|
| `VITE_API_BASE_URL` | `https://api-dev.tepidolacuenta.site` |
| `VITE_WS_BASE_URL` | `api-dev.tepidolacuenta.site` |
| `VITE_SENTRY_ENV` | `staging` |
| `VITE_SENTRY_DSN` | (mismo DSN del front) |

> Las env de Vite son build-time: cada push a `staging` rebuildea el preview con
> estos valores. Si querés un dominio fijo, asignar `app-dev.tepidolacuenta.site`
> a la rama `staging` desde Vercel → Domains.

---

## 6. Google OAuth

Agregar el redirect URI de staging a la consola de Google Cloud
(APIs & Services → Credentials → OAuth client → Authorized redirect URIs):

```
https://api-dev.tepidolacuenta.site/api/v1/auth/google/callback
```

---

## 7. MercadoPago (test)

En el panel de MercadoPago Developers, usar las **credenciales de prueba** y
configurar la URL de webhook de test apuntando a
`https://api-dev.tepidolacuenta.site/api/v1/payments/webhook`. Usar
[usuarios de prueba](https://www.mercadopago.com.ar/developers/es/docs/checkout-pro/additional-content/your-integrations/test/accounts)
para simular pagos sin mover dinero real.

---

## Flujo de trabajo

```bash
# 1. Crear la rama staging (una sola vez)
git checkout main && git pull
git checkout -b staging && git push -u origin staging

# 2. Para probar algo: mergear/pushear a staging
git checkout staging
git merge <rama-feature>
git push                      # dispara deploy-staging.yml

# 3. Validar en api-dev / preview de Vercel
# 4. Cuando está OK, recién ahí va a main (producción)
```

> El workflow `deploy-staging.yml` debe existir en la rama `staging` para que el
> trigger por push funcione. Asegurate de que esté mergeado ahí.
