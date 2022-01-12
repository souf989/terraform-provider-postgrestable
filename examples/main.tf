terraform {
  required_providers {
    postgrestable = {
      version = "0.1"
      source = "local.com/test/postgrestable"
    }
  }
}

provider "postgrestable" {
  database = "rds"
  scheme  = "postgres"
  host = "localhost"
  port = 5432
  username = "rds"
  password = "rds"
  sslmode  = "disable"
}

resource "postgrestable_table" "first_table" {
  table = "t_lien_affaire_responsable"
  schema = "public"
  columns {
    name = "dk_code_affaire"
    type = "varchar"
  }
  columns {
    name = "dk_code_responsable"
    type = "varchar"
  }
  columns {
    name = "date_debut"
    type = "varchar"
  }
  columns {
    name = "date_fin"
    type = "varchar"
  }
  columns {
    name = "id_editeur"
    type = "numeric"
  }
  columns {
    name = "ref_traitement"
    type = "varchar"
  }
  columns {
    name = "date_traitement"
    type = "timestamp"
  }

}