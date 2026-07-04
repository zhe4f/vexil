package config

import (
	"vexil/internal/protocol"
)

// Snapshot 创建当前全局配置的快照
func Snapshot() protocol.TransferConfig {
	return protocol.TransferConfig{
		NumConns:           NumConns,
		WindowSize:         WindowSize,
		MaxChunk:           MaxChunk,
		DialTimeout:        DialTimeout,
		ReadTimeout:        ReadTimeout,
		AcceptTimeout:      AcceptTimeout,
		CompleteTimeout:    CompleteTimeout,
		SHA256Timeout:      SHA256Timeout,
		StallInitial:       StallInitialTolerance,
		StallInterval:      StallCheckInterval,
		StallThreshold:     StallThreshold,
		ACKInterval:        ACKInterval,
		ResumeSaveInterval: ResumeSaveInterval,
		DrainRetries:       DrainRetries,
		DrainRetryTimeout:  DrainRetryTimeout,
		DialRetries:        DialRetries,
		DialRetryBaseDelay: DialRetryBaseDelay,
		ConnEstablishGap:   ConnEstablishGap,
		CompleteSendWait:   CompleteSendWait,
		TCPReadBufSize:     TCPReadBufSize,
		TCPWriteBufSize:    TCPWriteBufSize,
		TrackerSamples:     TrackerSampleSize,
		TrackerInterval:    TrackerUpdateInterval,

		// TLS 配置
		TLSEnabled:            TLSEnabled,
		TLSCertFile:           TLSCertFile,
		TLSKeyFile:            TLSKeyFile,
		TLSInsecureSkipVerify: TLSInsecureSkipVerify,
		TLSSessionCacheSize:   TLSSessionCacheSize,
	}
}