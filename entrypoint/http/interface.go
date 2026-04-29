package entrypointHttp

type Config interface {
	HttpHost() string
	HttpPort() string
}
