package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
)

// 这里有一份指导 https://tonybai.com/2015/04/30/go-and-https/

// 使用其他工具测试命令
// openssl s_client -connect 127.0.0.1:8100
// 这个为什么也能连接，奇怪， 命令输入之后，要输入 http Get
// GET / HTTP/1.1
// Host: 127.0.0.1:8100
// 就好像是在忽略证书校验，直接访问了
// openssl s_client -CAfile ca.crt -connect 127.0.0.1:8100
// 效果跟上面的命令一样

// curl https://127.0.0.1:8100
// 报错 curl: (60) Peer's Certificate issuer is not recognized.

// curl --cacert ca.crt https://127.0.0.1:8100
// 访问通过

// tls 校验地址不通过
// 报错 x509: certificate is valid for 192.168.28.106, 192.168.20.234, 192.168.0.1, 127.0.0.1, not x.x.x.x
// 绕过办法是设置 tls.Config ServerName 为上述一个有效的值，这样就使校验合法
// 看到 ServerName 使用的代码位置是
// net/http/transport.go:1488
// func (pconn *persistConn) addTLS(name string, trace *httptrace.ClientTrace) error
// 这个地址是默认从 https path 里取的
// 用这个命令查询
// " X509v3 Subject Alternative Name:"
// openssl s_client -connect 10.33.111.6:443  | openssl x509 -noout -text

// SAN 决定了允许以什么样的 http path 访问该服务器
// 其他 path 访问会报告错误

// NewCertPool read ca.cert files to make CertPool.
func NewCertPool(CAFiles []string) (*x509.CertPool, error) {
	cp := x509.NewCertPool()
	for _, CAFile := range CAFiles {
		pemByte, err := os.ReadFile(CAFile)
		if err != nil {
			return nil, err
		}
		ok := cp.AppendCertsFromPEM(pemByte)
		if !ok {
			return nil, fmt.Errorf("failed AppendCertsFromPEM() for %v", CAFile)
		}
	}
	return cp, nil
}

func makeTlsConfig() *tls.Config {
	cp, err := NewCertPool([]string{"ca.cert"})
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		InsecureSkipVerify: false,
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
		},
		RootCAs: cp,
		// ClientCAs: // 配置在 server 端，用来验证 client
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}
}

// 验证结论： 做到单向认证 Server 配置 server.cert server.key 文件
// Client 配置 RootCAs: ca.cert 用来通过 server.cert 的校验
// 如果是双向认证，则 Server 的 tls.Config 要配置 ClientCAs

func http1() {
	addr := "https://127.0.0.1:8100"
	trans0 := http.DefaultTransport.(*http.Transport)
	trans1 := trans0.Clone()
	trans1.TLSClientConfig = makeTlsConfig()
	clt := http.Client{
		Transport: trans1,
	}
	resp, err := clt.Get(addr)
	if err != nil {
		panic(err)
	}
	content, _ := httputil.DumpRequest(resp.Request, true)
	fmt.Printf("%s\n", content)
	content, _ = httputil.DumpResponse(resp, true)
	fmt.Printf("%s\n", content)
}

func main() {
	http1()
	fmt.Printf("main exit\n")
}
