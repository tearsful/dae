/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2022-2025, daeuniverse Organization <dae@v2raya.org>
 */

package outbound

import (
	_ "github.com/tearsful/outbound/dialer/http"
	_ "github.com/tearsful/outbound/dialer/hysteria2"
	_ "github.com/tearsful/outbound/dialer/juicity"
	_ "github.com/tearsful/outbound/dialer/shadowsocks"
	_ "github.com/tearsful/outbound/dialer/shadowsocksr"
	_ "github.com/tearsful/outbound/dialer/socks"
	_ "github.com/tearsful/outbound/dialer/trojan"
	_ "github.com/tearsful/outbound/dialer/tuic"
	_ "github.com/tearsful/outbound/dialer/v2ray"
	_ "github.com/tearsful/outbound/protocol/hysteria2"
	_ "github.com/tearsful/outbound/protocol/juicity"
	_ "github.com/tearsful/outbound/protocol/shadowsocks"
	_ "github.com/tearsful/outbound/protocol/trojanc"
	_ "github.com/tearsful/outbound/protocol/tuic"
	_ "github.com/tearsful/outbound/protocol/vless"
	_ "github.com/tearsful/outbound/protocol/vmess"
	_ "github.com/tearsful/outbound/transport/simpleobfs"
	_ "github.com/tearsful/outbound/transport/tls"
	_ "github.com/tearsful/outbound/transport/ws"
)
