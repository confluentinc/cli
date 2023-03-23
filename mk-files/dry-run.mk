define dry-run
$(if $(DRY_RUN),echo [DRY RUN] $(1),$(1))
endef