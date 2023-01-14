deploy:
	helm upgrade murmurations-allocator ./murmurationsSynchronizer --set env=$(DEPLOY_ENV) --install --atomic