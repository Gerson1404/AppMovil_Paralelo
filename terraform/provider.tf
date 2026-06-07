terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  # Usaremos la versión local del estado por ahora. 
  # En producción real se suele guardar en S3.
}

provider "aws" {
  region = var.aws_region
}