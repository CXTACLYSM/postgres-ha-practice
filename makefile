# Выбор инфры: make env INFRA=single (default) или make env INFRA=ha
INFRA ?= single

# Load environment file
ifneq (,$(wildcard ./$(ENV_INFRA_DIR)/.$(INFRA).env))
    include $(ENV_INFRA_DIR)/.$(INFRA).env
    export
endif

# Colors for output
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RED    := \033[0;31m
NC     := \033[0m # No Color

APP_SERVICES := target

ENV_DIR := .envs
ENV_APP_DIR := $(ENV_DIR)/app
ENV_INFRA_DIR := $(ENV_DIR)/infra
ENV_APP_SERVICES_DIR := $(ENV_APP_DIR)/services
ENV_INFRA_SERVICES_DIR := $(ENV_INFRA_DIR)/services
ENV_APP_TEMPLATES_DIR := $(ENV_APP_DIR)/templates
ENV_INFRA_TEMPLATES_DIR := $(ENV_INFRA_DIR)/templates

# Docker compose with env file
DC := docker compose -f compose.$(INFRA).yml --env-file $(ENV_INFRA_DIR)/.$(INFRA).env

VALID_INFRA := single ha

.PHONY: env
env:
	@if ! echo "$(VALID_INFRA)" | grep -wq "$(INFRA)"; then \
        echo "$(RED)Unknown INFRA='$(INFRA)'. Valid values: $(VALID_INFRA)$(NC)"; \
        exit 1; \
    fi
	@echo "$(GREEN)Generating service env files from templates...$(NC)"
	@set -a && . $(ENV_INFRA_DIR)/.$(INFRA).env && set +a && \
		for svc in $(APP_SERVICES); do \
			mkdir -p $(ENV_APP_SERVICES_DIR)/$$svc; \
			if [ -f $(ENV_APP_TEMPLATES_DIR)/.env.$$svc.template ]; then \
				envsubst < $(ENV_APP_TEMPLATES_DIR)/.env.$$svc.template > $(ENV_APP_SERVICES_DIR)/$$svc/.env; \
				echo "  $(GREEN)✓$(NC) app/services/$$svc/.env"; \
			else \
				echo "  $(YELLOW)⚠$(NC) app/services/$$svc/.env — template not found, skipped"; \
			fi; \
		done && \
		for svc in $(INFRA_SERVICES); do \
			mkdir -p $(ENV_INFRA_SERVICES_DIR)/$$svc; \
			if [ -f $(ENV_INFRA_TEMPLATES_DIR)/.env.$$svc.template ]; then \
				envsubst < $(ENV_INFRA_TEMPLATES_DIR)/.env.$$svc.template > $(ENV_INFRA_SERVICES_DIR)/$$svc/.env; \
				echo "  $(GREEN)✓$(NC) infra/services/$$svc/.env"; \
			else \
				echo "  $(YELLOW)⚠$(NC) infra/services/$$svc/.env — template not found, skipped"; \
			fi; \
		done
	@echo "$(GREEN)All service env files generated!$(NC)"

.PHONY: docker-build
docker-build:
	docker build -f builds/Dockerfile -t postgres-ha-practice:1.0 --target app .

.PHONY: up
up: env docker-build ## Start all services
	@echo "$(GREEN)Starting all services...$(NC)"
	$(DC) up -d --build
	@echo "$(GREEN)Services started!$(NC)"