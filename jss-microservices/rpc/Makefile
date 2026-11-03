PROTOC       = protoc
PROTO_DIR    = rpc
OUT_DIR      = gen

# Lấy danh sách proto, bỏ qua google/api
PROTO_FILES := $(shell find $(PROTO_DIR) -name "*.proto" ! -path "$(PROTO_DIR)/google/*" ! -path "$(PROTO_DIR)/protoc-gen-openapiv2/*")

proto: clean
	@mkdir -p $(OUT_DIR)
	$(PROTOC) \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(OUT_DIR) --grpc-gateway_opt=paths=source_relative \
		$(PROTO_FILES)
clean:
	@echo "🧹 Cleaning generated files..."
	@rm -rf $(OUT_DIR)

update:
	go mod tidy 
	go mod vendor

push:
	git add .
	git commit -m 'refactor: update package'
	git push
	git hash

.PHONY: proto
