variable "aws_region" {
  description = "Región de AWS donde se desplegará la infraestructura"
  type        = string
  default     = "us-east-1"
}

variable "database_url" {
  description = "URL de conexión a la base de datos Neon"
  type        = string
  sensitive   = true
}

variable "jwt_secret" {
  description = "Secreto para firmar los tokens JWT"
  type        = string
  sensitive   = true
}