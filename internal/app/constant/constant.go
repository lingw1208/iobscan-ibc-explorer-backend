package constant

const (
	EnvNameConfigFilePath = "CONFIG_FILE_PATH"

	DefaultTimezone       = "UTC"
	DefaultTimeFormat     = "2006-01-02 15:04:05"
	DateFormat            = "2006-01-02"
	TimeFormatMMDD        = "0102"
	DefaultCurrency       = "$"
	UnknownTokenPrice     = -1
	UnknownDenomAmount    = "-1"
	ZeroDenomAmount       = "0"
	IBCTokenPrefix        = "ibc"
	IBCHopsIndex          = "/channel"
	DefaultValuePrecision = 5
	ChannelStateOpen      = "STATE_OPEN"
	DefaultPageSize       = 10
	DefaultPageNum        = 1
	OtherDenom            = "others"
	AllChain              = "allchain"
	Cosmos                = "cosmos"
	ChainNameCosmosHub    = "cosmoshub"
	DenomAtom             = "uatom"
	Iris                  = "iris"
	PortTransfer          = "transfer"
	DefaultUnboundTime    = 1209600

	IncreaseSymbol = "+"
	DecreaseSymbol = "-"

	DefaultLimit = 500
	IncreHeight  = 5000

	DisplayIbcRecordMax = 500000

	MsgTypeTransfer           = "transfer"
	MsgTypeRecvPacket         = "recv_packet"
	MsgTypeTimeoutPacket      = "timeout_packet"
	MsgTypeAcknowledgement    = "acknowledge_packet"
	MsgTypeUpdateClient       = "update_client"
	MsgTypeChannelOpenConfirm = "channel_open_confirm"

	ChannelOpenStatisticName  = "channel_opened"
	ChannelCloseStatisticName = "channel_closed"
	ChannelAllStatisticName   = "channel_all"
	Channel24hStatisticName   = "channels_24hr"
	Chains24hStatisticName    = "chains_24hr"
	ChainsAllStatisticName    = "chain_all"
	Tx24hAllStatisticName     = "tx_24hr_all"
	TxLatestAllStatisticName  = "tx_latest_all"
	TxAllStatisticName        = "tx_all"
	TxSuccessStatisticName    = "tx_success"
	TxFailedStatisticName     = "tx_failed"
	BaseDenomAllStatisticName = "base_denom_all"
	DenomAllStatisticName     = "denom_all"
	RelayersStatisticName     = "relayers"

	IBCConnectionChainsIconUri = "https://iobscan.io/resources/home/connection-chains/%s.png"

	ETHSECP256K1 = "ethsecp256k1"
	SECP256K1    = "secp256k1"
	ICS20        = "ics20"

	ChainFlowTrendDays = 365

	ExportTxsNum = 1000
)

var HomeStatistics = []string{
	ChannelOpenStatisticName, ChannelCloseStatisticName, ChannelAllStatisticName, Channel24hStatisticName,
	Chains24hStatisticName, ChainsAllStatisticName,
	Tx24hAllStatisticName, TxAllStatisticName, TxSuccessStatisticName, TxFailedStatisticName,
	BaseDenomAllStatisticName, DenomAllStatisticName,
}

var RelayerDetailTxsType = []string{MsgTypeRecvPacket, MsgTypeAcknowledgement, MsgTypeTimeoutPacket}

const (
	UnAuth = "Others"
	//AllChain = "allchain"
)

const (
	//packet没有过期且没有发现成功的RecvPacket
	NoFoundSuccessRecvPacket = "NoFoundSuccessRecvPacket"

	//packet没有过期且没有发现成功的AcknowledgePacket
	NoFoundSuccessAcknowledgePacket = "NoFoundSuccessAcknowledgePacket"
	//packet已过期
	NoFoundSuccessTimeoutPacket = "NoFoundSuccessTimeoutPacket"
	//dc_chain_id为空,历史setting未处理的数据
	NoFoundDcChain = "NoFoundDcChainId"
)
