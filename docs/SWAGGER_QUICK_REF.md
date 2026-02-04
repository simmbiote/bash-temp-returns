# 📚 Swagger Documentation - Quick Reference

## 🌐 Access Points

| Resource | URL |
|----------|-----|
| **Swagger UI** | http://localhost:8080/swagger/index.html |
| **JSON Spec** | http://localhost:8080/swagger/doc.json |
| **YAML Spec** | docs/swagger.yaml |

## 📋 Documented Endpoints (7)

### Conversations (4 endpoints)

```
POST   /api/v1/conversations              Create conversation
GET    /api/v1/conversations/{id}         Get conversation
GET    /api/v1/conversations/{id}/messages Get messages  
POST   /api/v1/conversations/{id}/messages Send message
```

### Returns (3 endpoints)

```
POST   /api/v1/returns     Create return request
GET    /api/v1/returns     List returns
GET    /api/v1/returns/{id} Get return details
```

## 🔧 Quick Commands

### Start Server
```bash
ORDERS_API_URL=mock go run cmd/api/main.go
```

### Regenerate Docs
```bash
~/go/bin/swag init -g cmd/api/main.go -o docs
```

### View in Browser
```bash
open http://localhost:8080/swagger/index.html
```

## 📦 Generated Files

```
docs/
├── docs.go          # Embedded Go documentation
├── swagger.json     # OpenAPI JSON spec
└── swagger.yaml     # OpenAPI YAML spec
```

## ✨ Features

- ✅ Interactive API testing
- ✅ Complete request/response schemas
- ✅ Model definitions
- ✅ Error responses documented
- ✅ Export to Postman/Insomnia
- ✅ Auto-generated from code

## 📖 Documentation

- Full Guide: [docs/SWAGGER.md](SWAGGER.md)
- Summary: [docs/SWAGGER_SUMMARY.md](SWAGGER_SUMMARY.md)

---

**Version**: 0.2.0 | **Coverage**: 100% (7/7 endpoints)
