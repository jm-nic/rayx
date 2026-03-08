[English](/README.md) | [فارسی](/README.fa_IR.md) | [العربية](/README.ar_EG.md) |  [中文](/README.zh_CN.md) | [Español](/README.es_ES.md) | [Русский](/README.ru_RU.md)

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./media/3x-ui-dark.png">
    <img alt="3x-ui" src="./media/3x-ui-light.png">
  </picture>
</p>

[![Release](https://img.shields.io/github/v/release/cortez62/3x-ui-uuid.svg)](https://github.com/cortez62/3x-ui-uuid/releases)
[![Build](https://img.shields.io/github/actions/workflow/status/cortez62/3x-ui-uuid/release.yml.svg)](https://github.com/cortez62/3x-ui-uuid/actions)
[![GO Version](https://img.shields.io/github/go-mod/go-version/cortez62/3x-ui-uuid.svg)](#)
[![License](https://img.shields.io/badge/license-GPL%20V3-blue.svg?longCache=true)](https://www.gnu.org/licenses/gpl-3.0.en.html)

## ¿Qué es 3x-ui-uuid?

**3x-ui-uuid** es un panel de control web de código abierto para gestionar servidores [Xray-core](https://github.com/XTLS/Xray-core). Es un fork del proyecto [3X-UI](https://github.com/MHSanaei/3x-ui) que agrega una **API pública de consulta de vencimiento por UUID** y mejoras en el sistema de enlaces de suscripción, facilitando el despliegue de servicios VPN/proxy para múltiples usuarios.

> [!IMPORTANT]
> Este proyecto es solo para uso personal y comunicación. No lo utilice para actividades ilegales ni en entornos de producción críticos.

---

## ¿Qué problemas resuelve?

### Para el administrador del servidor
Un administrador que quiere ofrecer acceso VPN/proxy a varios usuarios necesita:
- Crear y gestionar múltiples clientes (con límites de tráfico, fechas de vencimiento y restricciones de IP).
- Distribuir configuraciones de forma segura sin exponer el panel completo.
- Monitorear el estado del servidor y el consumo de tráfico de cada cliente.
- Recibir alertas y gestionar usuarios desde cualquier lugar, incluso desde Telegram.

**3x-ui-uuid** resuelve todo esto con un panel web intuitivo, un servidor de suscripciones independiente y un bot de Telegram integrado.

### Para el usuario final
Un usuario que recibe acceso VPN necesita:
- Importar fácilmente su configuración en su cliente Xray (v2rayN, v2rayNG, Shadowrocket, Nekoray, etc.).
- Saber cuándo vence su cuenta o cuánto tráfico le queda, sin necesidad de acceder al panel.

**3x-ui-uuid** resuelve esto mediante **enlaces de suscripción personalizados por SubID** (actualizables automáticamente por el cliente) y una **API pública de vencimiento por UUID** que no requiere autenticación.

---

## Características principales

### Panel web
- Interfaz gráfica completa accesible desde el navegador (puerto `2053` por defecto).
- Gestión de múltiples **inbounds** (entradas/servidores) con soporte para los protocolos:
  - VMess, VLESS, Trojan, Shadowsocks, SOCKS, HTTP y WireGuard.
- Configuración de transporte: TCP, WebSocket, HTTP/2, gRPC, QUIC, KCP, HTTPUpgrade y SplitHTTP.
- Soporte de seguridad: TLS, XTLS, Reality.
- Panel de estadísticas en tiempo real (CPU, memoria, red, estado de Xray).
- Autenticación con soporte de doble factor (2FA).
- Integración con LDAP para autenticación corporativa.

### Gestión de clientes
- Creación individual o masiva de clientes con:
  - UUID único por cliente.
  - **SubID** (identificador de suscripción) generado aleatoriamente.
  - Límite de tráfico total (en GB).
  - Fecha de vencimiento configurable.
  - Límite de IPs simultáneas.
  - Vinculación con ID de Telegram para notificaciones personalizadas.
- Reinicio de tráfico manual o programado (diario, semanal, mensual).
- Habilitación/deshabilitación individual de clientes.

### Sistema de suscripciones
El servidor de suscripciones corre de forma independiente (puerto `2096` por defecto) y ofrece:
- **Enlace de suscripción estándar** (`/sub/<SubID>`): devuelve todos los inbounds habilitados del cliente en formato base64, compatible con la mayoría de los clientes Xray.
- **Enlace de suscripción JSON** (`/json/<SubID>`): devuelve la configuración en formato JSON extendido, con soporte de fragmentación, multiplexación y ruido.
- **Página web de suscripción**: muestra al usuario su tráfico usado, tráfico restante, fecha de vencimiento y última conexión.
- **Códigos QR**: generación y descarga de QR para cada enlace de suscripción.
- Soporte opcional de cifrado y validación de dominio para mayor seguridad.

### API pública de vencimiento por UUID ⭐ (característica única de este fork)
Endpoint sin autenticación que permite a cualquier cliente consultar el estado de su cuenta usando solo su UUID:

```
GET /panel/api/public/client-expiry/:uuid
```

Respuesta de ejemplo:
```json
{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "expiryTime": 1735689600000,
  "expiryDate": "2025-01-01",
  "daysRemaining": 10,
  "expired": false
}
```

Esto permite integrar la consulta de vencimiento en bots, aplicaciones móviles o páginas informativas sin exponer credenciales del panel.

### Bot de Telegram
- Comandos disponibles:
  - `/start` — Bienvenida e información del servidor.
  - `/status` — Estado del servidor (CPU, memoria, uptime, Xray).
  - `/usage [email]` — Consulta de tráfico de un cliente.
  - `/id` — Obtener el ID de Telegram propio.
  - `/inbound` — Búsqueda de inbounds (solo administradores).
  - `/restart` — Reinicio del servicio Xray (solo administradores).
- Gestión completa de clientes desde el chat (agregar, editar, eliminar, resetear tráfico, cambiar vencimiento).
- Envío de enlaces de suscripción, links individuales y códigos QR directamente al usuario.
- Notificaciones automáticas de vencimiento y límite de tráfico.
- Envío de backups de la base de datos a los administradores.
- Notificaciones programadas por cron.

### Seguridad y administración
- SSL/TLS para el panel web y el servidor de suscripciones.
- Gestión de certificados integrada.
- Seguimiento de IPs por cliente para prevenir uso compartido no autorizado.
- Archivos de log separados para el panel y las IPs baneadas.

---

## Diferencias con el 3X-UI original

| Característica | 3X-UI original | 3x-ui-uuid |
|---|---|---|
| API pública de vencimiento por UUID | ❌ No disponible | ✅ `GET /panel/api/public/client-expiry/:uuid` |
| Consulta de vencimiento sin autenticación | ❌ | ✅ Solo con el UUID del cliente |
| Badges e instalación apuntando al fork | ❌ | ✅ Ajustados a `cortez62/3x-ui-uuid` |
| Script de instalación personalizado | Repo original | ✅ `install.sh` de este repositorio |

El resto de las funcionalidades (panel web, suscripciones, bot de Telegram, soporte de protocolos) son compartidas con el proyecto base y se mantienen actualizadas.

---

## Instalación rápida

```bash
bash <(curl -Ls https://raw.githubusercontent.com/cortez62/3x-ui-uuid/main/install.sh)
```

Tras la instalación, el panel estará disponible en `http://<IP>:2053` con las credenciales por defecto `admin` / `admin` (cámbielas inmediatamente).

### Con Docker

```bash
docker compose up -d
```

O con imagen preconstruida:

```bash
docker run -itd \
  -p 2053:2053 \
  -p 2096:2096 \
  -v $PWD/db/:/etc/x-ui/ \
  -v $PWD/cert/:/root/cert/ \
  --name 3x-ui \
  ghcr.io/cortez62/3x-ui-uuid:latest
```

---

## Gestión del panel (x-ui.sh)

Después de la instalación, el script `x-ui` queda disponible en el sistema:

```
x-ui              # Menú interactivo
x-ui start        # Iniciar el panel
x-ui stop         # Detener el panel
x-ui restart      # Reiniciar el panel
x-ui status       # Ver estado del servicio
x-ui log          # Ver logs en tiempo real
x-ui update       # Actualizar el panel
```

---

## Acceso predeterminado

| Recurso | Valor por defecto |
|---|---|
| Puerto del panel | `2053` |
| Puerto de suscripciones | `2096` |
| Ruta del panel | `/` |
| Ruta de suscripción | `/sub/` |
| Ruta JSON | `/json/` |
| Usuario | `admin` |
| Contraseña | `admin` |

---

## Documentación adicional

Para documentación completa sobre configuración avanzada, visita la [Wiki del proyecto 3X-UI](https://github.com/MHSanaei/3x-ui/wiki).

---

## Reconocimientos

- [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui) — proyecto base en el que se basa este fork.
- [alireza0](https://github.com/alireza0/) — contribuciones al proyecto original.
- [Iran v2ray rules](https://github.com/chocolate4u/Iran-v2ray-rules) (Licencia: **GPL-3.0**): _Reglas de enrutamiento mejoradas para v2ray/xray con dominios iraníes y enfoque en seguridad._
- [Russia v2ray rules](https://github.com/runetfreedom/russia-v2ray-rules-dat) (Licencia: **GPL-3.0**): _Reglas de enrutamiento V2Ray actualizadas para dominios y direcciones bloqueadas en Rusia._

---

## Licencia

Este proyecto se distribuye bajo la licencia [GPL v3](https://www.gnu.org/licenses/gpl-3.0.en.html).
