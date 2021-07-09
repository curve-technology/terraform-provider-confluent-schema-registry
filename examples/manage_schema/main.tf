terraform {
  required_providers {
    schemaregistry = {
      source = "curve-technology/confluent-schema-registry"
      version = "~> 0.1.0"
    }
  }
}

variable "schemas_mapping" {
  description = "Schemas mapping between Subject (topic name) and proto schema"
  type = list(object({
    subject    = string
    references = list(string)
    schema     = string
  }))
  default = [
    {
subject    = "topic1-value"
references = []
schema     = <<-EOT
syntax = "proto3";
package messaging.domain1.v1;

message Message1 {
  string field1 = 1;
  string field2 = 2;
  string field3 = 3;
  string field4 = 4;
  string field5 = 5;
}
EOT
    }
  ]
}


provider "schemaregistry" {
  url = "http://localhost:8081"
}


resource "schemaregistry_bulk_dependent_schemas" "main" {

  dynamic "schemas_mapping" {
    for_each = var.schemas_mapping
    content {
      subject    = schemas_mapping.value["subject"]
      # references = schemas_mapping.value["references"]
      schema     = schemas_mapping.value["schema"]
    }
  }
}
