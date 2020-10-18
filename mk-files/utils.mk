define print-boxed-message
	@echo
	$(eval length=$(call string-length,$1))	
	$(eval box_len=$(shell echo $$(expr $(length) + 8)))
	$(call print-box-edge,$(box_len))
	@echo "    $1"
	$(call print-box-edge,$(box_len))
	@echo
endef

define string-length
	$(shell var=$1 && echo $${#var})
endef

define print-box-edge
	@printf '=%.0s' {1..$1}; echo
endef
