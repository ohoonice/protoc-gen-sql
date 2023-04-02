.PHONY: genproto
genproto:
	@protoc -Iproto --go_out=proto ohoonice/sql/sql.proto