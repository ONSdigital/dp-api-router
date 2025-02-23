job "dp-api-router" {
  datacenters = ["eu-west-1"]
  region      = "eu"
  type        = "system"
  priority    = 100

  meta {
    job_type = "system"
  }

  update {
    stagger          = "60s"
    min_healthy_time = "30s"
    healthy_deadline = "2m"
    max_parallel     = 1
    auto_revert      = true
  }

  group "web" {
    constraint {
      attribute = "${node.class}"
      value     = "web"
    }

    restart {
      attempts = 3
      delay    = "15s"
      interval = "1m"
      mode     = "delay"
    }

    task "dp-api-router-web" {
      driver = "docker"

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/{{DEPLOYMENT_BUCKET}}/dp-api-router/{{REVISION}}.tar.gz"
      }

      config {
        command = "${NOMAD_TASK_DIR}/start-task"

        args = ["./dp-api-router"]
       	
        image = "{{ECR_URL}}:concourse-{{REVISION}}"

        port_map {
          http = "${NOMAD_PORT_http}"
        }
      }

      service {
        name = "dp-api-router"
        port = "http"
        tags = ["web", "system_job"]
        check {
          type     = "http"
          path     = "/health"
          interval = "10s"
          timeout  = "2s"
        }
      }

      resources {
        cpu    = "{{WEB_RESOURCE_CPU}}"
        memory = "{{WEB_RESOURCE_MEM}}"

        network {
          port "http" {}
        }
      }

      template {
        source      = "${NOMAD_TASK_DIR}/vars-template"
        destination = "${NOMAD_TASK_DIR}/vars"
      }

      # Create a config file for deprecations with content from additional secrets
      template {
        data        = <<EOH
        {{ with secret (print "secret/" (env "NOMAD_TASK_NAME") "-additional") }}{{ toJSON .Data.deprecations }}{{ end }}
        EOH
        destination = "${NOMAD_SECRETS_DIR}/deprecations.json"
        env         = false
      }

      vault {
        policies = ["dp-api-router-web"]
      }
    }
  }

  group "publishing" {
    constraint {
      attribute = "${node.class}"
      value     = "publishing"
    }

    restart {
      attempts = 3
      delay    = "15s"
      interval = "1m"
      mode     = "delay"
    }

    task "dp-api-router-publishing" {
      driver = "docker"

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/{{DEPLOYMENT_BUCKET}}/dp-api-router/{{REVISION}}.tar.gz"
      }

      config {
        command = "${NOMAD_TASK_DIR}/start-task"

        args = ["./dp-api-router"]
       	
        image = "{{ECR_URL}}:concourse-{{REVISION}}"

        port_map {
          http = "${NOMAD_PORT_http}"
        }
      }

      service {
        name = "dp-api-router"
        port = "http"
        tags = ["publishing", "system_job"]
        check {
          type     = "http"
          path     = "/health"
          interval = "10s"
          timeout  = "2s"
        }
      }

      resources {
        cpu    = "{{PUBLISHING_RESOURCE_CPU}}"
        memory = "{{PUBLISHING_RESOURCE_MEM}}"

        network {
          port "http" {}
        }
      }

      template {
        source      = "${NOMAD_TASK_DIR}/vars-template"
        destination = "${NOMAD_TASK_DIR}/vars"
      }

      vault {
        policies = ["dp-api-router-publishing"]
      }
    }
  }

}
