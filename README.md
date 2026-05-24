# Broadside

> *A broadside: a single-sheet political pamphlet, printed cheaply and distributed widely. The original mass-communication tool.*

**A self-hosted email platform with OIDC SSO, for organizations that want to own their stack.**

Broadside is a fork of [Notifuse](https://github.com/Notifuse/notifuse) (AGPLv3) with first-class support for OIDC single sign-on (Authentik, Keycloak, Okta, etc.), built for organizations that need to run their own email/marketing stack rather than rent SaaS.

**The open-source alternative to Mailchimp, Brevo, Mailjet, Listmonk, Mailerlite, Klaviyo, Loop.so, etc.**

Built with Go and React. Same core engine as upstream Notifuse, with the SSO feature upstream chose not to ship.

<img alt="Email Editor" src="https://github.com/user-attachments/assets/f650ac1b-58fd-44fb-884d-e9811255f1e4" />

## 🚀 Key Features

### 📧 Email Marketing

- **Visual Email Builder**: Drag-and-drop editor with MJML components and real-time preview
- **Campaign Management**: Create, schedule, and send targeted email campaigns
- **A/B Testing**: Optimize campaigns with built-in testing for subject lines, content, and send times
- **List Management**: Advanced subscriber segmentation and list organization
- **Contact Profiles**: Rich contact management with custom fields and detailed profiles

### 🔧 Developer-Friendly

- **Easy Setup**: Interactive setup wizard for quick deployment and configuration
- **Transactional API**: Powerful REST API for automated email delivery
- **Webhook Integration**: Real-time event notifications and integrations
- **Liquid Templating**: Dynamic content with variables like `{{ contact.first_name }}`
- **Multi-Provider Support**: Connect with Amazon SES, Mailgun, Postmark, Mailjet, SparkPost, and SMTP

### 📊 Analytics & Insights

- **Open & Click Tracking**: Detailed engagement metrics and campaign performance
- **Real-time Analytics**: Monitor delivery rates, opens, clicks, and conversions
- **Campaign Reports**: Comprehensive reporting and analytics dashboard

### 🎨 Advanced Features

- **S3 File Manager**: Integrated file management with CDN delivery
- **Notification Center**: Centralized notification system for your applications
- **Responsive Templates**: Mobile-optimized email templates
- **Custom Fields**: Flexible contact data management
- **Workspace Management**: Multi-tenant support for teams and agencies

## 🏗️ Architecture

Broadside follows clean architecture principles with clear separation of concerns:

### Backend (Go)

- **Domain Layer**: Core business logic and entities (`internal/domain/`)
- **Service Layer**: Business logic implementation (`internal/service/`)
- **Repository Layer**: Data access and storage (`internal/repository/`)
- **HTTP Layer**: API handlers and middleware (`internal/http/`)

### Frontend (React)

- **Console**: Admin interface built with React, Ant Design, and TypeScript (`console/`)
- **Notification Center**: Embeddable widget for customer notifications (`notification_center/`)

### Database

- **PostgreSQL**: Primary data storage with Squirrel query builder

## 📁 Project Structure

```
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── domain/            # Business entities and logic
│   ├── service/           # Business logic implementation
│   ├── repository/        # Data access layer
│   ├── http/              # HTTP handlers and middleware
│   └── database/          # Database configuration
├── console/               # React-based admin interface
├── notification_center/   # Embeddable notification widget
├── pkg/                   # Public packages
└── config/                # Configuration files
```

## 📚 Documentation

Upstream Notifuse documentation applies to most features: **[docs.notifuse.com](https://docs.notifuse.com)**. Broadside-specific documentation (OIDC setup, group mappings) lives in this repo.

## 🔀 Relationship to Notifuse

Broadside is a hard fork of [Notifuse](https://github.com/Notifuse/notifuse). We track upstream by force-rebasing `main` onto upstream when syncing, then re-applying our additions on top. We do not contribute back upstream because the upstream contributor agreement requires full IP assignment — we prefer to keep our work under AGPLv3 ownership rather than transferred.

If you want the original, go to [github.com/Notifuse/notifuse](https://github.com/Notifuse/notifuse). If you want SSO + self-hosting without the SaaS upsell, you're in the right place.

## 📄 License

Broadside is released under the [GNU Affero General Public License v3.0](LICENCE.md), the same license as upstream Notifuse. Original Notifuse code remains copyright (C) 2025 Notifuse. Broadside additions retain that copyright and add ours.

## 🌟 Why Choose Broadside?

- **🪪 SSO first-class**: OIDC built in, with IdP group → workspace permission mapping
- **💰 Cost-Effective**: Self-hosted, no per-email pricing
- **🔒 Privacy-First**: Your data stays on your infrastructure
- **🛠️ Customizable**: Open-source (AGPLv3), extensible
- **📈 Scalable**: Built to handle millions of emails
- **🔧 Developer-Friendly**: Comprehensive API and webhook support
