terraform {
  required_providers {
    protogit = {
      source = "curve-technology/protogit"
      version = "~> 0.1.0"
    }

    schemaregistry = {
      source = "curve-technology/confluent-schema-registry"
      version = "~> 0.1.0"
    }
  }
}


provider "protogit" {
  url         = "github.com/curve-technology/terraform-provider-protogit"
  tag_version = "v0.1.0"
  password    = var.git_password
}


data "protogit_schemas" "schemas_collection" {
  entries {
    topic    = "topic1"
    section  = "value"
    filepath = "messaging/domain1/v1/event1.proto"
  }

  entries {
    topic    = "topic2"
    section  = "value"
    filepath = "messaging/domain2/v1/event1.proto"
  }
}


provider "schemaregistry" {
  url = "http://localhost:8081"
}


# output "protogit_output" {
#   value = data.protogit_schemas.schemas_collection
# }


resource "schemaregistry_bulk_dependent_schemas" "main" {

  dynamic "schemas_mapping" {
    for_each = data.protogit_schemas.schemas_collection.schemas_mapping
    content {
      subject = schemas_mapping.value.subject
      schema  = schemas_mapping.value.schema

      dynamic "references" {
        for_each = schemas_mapping.value.references
        content {
          subject = references.value
        }
      }
    }
  }
}
