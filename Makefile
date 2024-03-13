DEPLOY_ENV ?= development

deploy:
	helm upgrade murmurations-synchronizer ./murmurationsSynchronizer --set env=$(DEPLOY_ENV) --install --atomic