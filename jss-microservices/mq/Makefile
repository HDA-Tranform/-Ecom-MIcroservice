PROTOC       = protoc
PROTO_DIR    = events
OUT_DIR      = events

# Lấy danh sách proto, bỏ qua google/api
PROTO_FILES := $(shell find $(PROTO_DIR) -name "*.proto" ! -path "$(PROTO_DIR)/google/api/*")

proto: 
	@mkdir -p $(OUT_DIR)
	$(PROTOC) \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(OUT_DIR) --grpc-gateway_opt=paths=source_relative \
		$(PROTO_FILES)

update:
	go mod tidy 
	go mod vendor

.PHONY: proto