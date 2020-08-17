package session

import (
	"github.com/overtalk/bgo/pkg/service/packet"
	"github.com/overtalk/bgo/pkg/service/tunnel"
	"net"

	"github.com/overtalk/bgo/pkg/service/route"
)

func BuildServeFunc(optAgent int, router *route.Router) func(net.Conn) {
	return nil
}

// AgentService an agent service
type AgentService struct {
	router *route.Router
}

func NewAgentService(router *route.Router) *AgentService {
	return &AgentService{router: router}
}

func (as *AgentService) handleAgentCmd(sess *tunnel.BackendSession, pack packet.Packet) {
	cmd := pack.GetCmd()
	switch cmd {
	case packet.CmdPing:
		sess.UpdatePing()
	default:
		// TODO: log
		//zaplog.S.Errorf("agent@%s: packet: %v, invalid cmd(%d)",
		//	sess.ClientAddr(), cmd)
	}
}

// handleAgentRequest handle the cmd request
func (this *AgentService) handleAgentRequest(sess *tunnel.BackendSession, req *tunnel.BackendRequest) {
	sess.AddRequest()
	defer func() {
		if err := recover(); err != nil {
			//TODO: add log
		}
		req.Free()
		sess.DoneRequest()
	}()

	// no need to decrypt the data from an agent server
	inPacket := req.GetPacket()
	if inPacket.IsCmdSize() || inPacket.IsCmdProto() {
		this.handleAgentCmd(sess, inPacket)
		return
	}

	connID := inPacket.GetConnID()
	//show packet content
	// TODO: add log
	//zaplog.S.Debugf("agent@%s: cid: %d, packet: %v, size: %d",
	//	sess.ClientAddr(), connID, inPacket, len(inPacket))
	clientRequest := NewRequestFromAgent(inPacket)

	result, isTimeout := this.router.Dispatch(clientRequest)
	if isTimeout {
		//TODO: add log
		//zaplog.S.Errorf(
		//	"agent@%s response timeout: cid: %d, mid: %d, aid: %d",
		//	sess.ClientAddr(), connID, clientRequest.MID, clientRequest.AID)
	}

	// TODO: log
	//zaplog.S.Debugf(
	//	"agent@%s response: cid: %d, mid: %d, aid: %d, out: [%v]",
	//	sess.ClientAddr(), connID, clientRequest.MID, clientRequest.AID, result,
	//)

	dataLoad, err := result.Marshal()
	if err != nil {
		//TODO: log
		//zaplog.S.Errorf(
		//	"agent@%s marshal error: cid: %d, mid: %d, aid: %d, err: %v",
		//	sess.ClientAddr(), connID, clientRequest.MID, clientRequest.AID, err)
		return
	}

	outPacket := packet.NewFromData(dataLoad, nil, packet.NoneCompresser)
	outPacket.SetConnID(connID)
	outPacket.SetProtoMID(inPacket.GetProtoMID())
	outPacket.SetProtoAID(inPacket.GetProtoAID())
	outPacket.SetProtoVer(inPacket.GetProtoVer())

	// zaplog.S.Debugf(
	//	"agent@%s response: cid: %d, mid: %d, aid: %d, out: %v",
	//	as.sess.ClientAddr(), connID, clientRequest.MID, clientRequest.AID, outPacket,
	// )

	_, err = sess.Write(outPacket)
	if err != nil {
		//TODO: log
		//zaplog.S.Errorf(
		//	"write agent@%s response: cid: %d, mid: %d, aid: %d, err: %v",
		//	sess.ClientAddr(), connID, clientRequest.MID,
		//	clientRequest.AID, zeroutil.ParseNetError(err))
	}

}

func (this *AgentService) Serve(nc net.Conn) {
	backendSess := tunnel.NewBackendSession(0, nc)
	defer func() {
		if err := recover(); err != nil {
			//TODO: log
		}
		backendSess.Close()
	}()

	// it's a long session
	backendSess.CheckPing()

	for {
		inReq, err := backendSess.ReadRequest()
		if err != nil {
			inReq.Free()
			// TODO: add log
		} else {
			go this.handleAgentRequest(backendSess, inReq)
		}
	}
}

// LocalAgentSession a local agent session
type LocalAgentService struct {
	router *route.Router
}

// NewLocalAgentService create a LocalAgentService struct
func NewLocalAgentService(router *route.Router) *LocalAgentService {
	return &LocalAgentService{router: router}
}

// Serve serve a tcp session from the frontend
func (as *LocalAgentService) Serve(nc net.Conn) {
	frontendSess := tunnel.NewFrontendSession(nc)
	defer func() {
		if err := recover(); err != nil {
			//TODO: log
			//zaplog.S.Error(err)
			//			zaplog.S.Error(zap.Stack("").String)
		}
		frontendSess.Close()
	}()

	inPacket, err := frontendSess.ReadPacket()
	if err != nil {
		// TODO: error handler
		//if e := errors.Cause(err); !zeroutil.IsNetTimeout(e) {
		//	zaplog.S.Errorf(
		//		"read client@%s request: %v", frontendSess.ClientAddr(), err)
		//	return
		//}
	}

	if !inPacket.IsValid() {
		//TODO: log
		//zaplog.S.Errorf("read client@%s request: %v, invalid packet",
		//	frontendSess.ClientAddr(), inPacket)
		return
	}

	// decrypt the packet and clear the FlagXOR,
	// and a game server doesn't decrypt it again
	inPacket.Decrypt(packet.XORCrypto)

	// cmd proto is not permited
	if inPacket.IsCmdSize() || inPacket.IsCmdProto() {
		//TODO: add log
		//zaplog.S.Errorf(
		//	"read client@%s request: %v, cmd(%d) is not permitted",
		//	frontendSess.ClientAddr(), inPacket, inPacket.GetCmd())
		return
	}

	sid := inPacket.GetConnID()
	// show packet content
	//zaplog.S.Debugf("client@%s: packet: %v, size: %d",
	//	frontendSess.ClientAddr(), inPacket, len(inPacket))

	clientRequest := NewRequestFromAgent(inPacket)
	result, isTimeout := as.router.Dispatch(clientRequest)
	if isTimeout {
		// TODO: log
		//	zaplog.S.Errorf(
		//		"client@%s response timeout: mid: %d, aid: %d",
		//		frontendSess.ClientAddr(), clientRequest.MID, clientRequest.AID)
	}

	//zaplog.S.Debugf(
	//	"agent@%s response: mid: %d, aid: %d, out: [%v]",
	//	frontendSess.ClientAddr(), clientRequest.MID, clientRequest.AID, result,
	//)

	dataload, err := result.Marshal()
	if err != nil {
		//TODO : log
		//zaplog.S.Errorf(
		//	"agent@%s marshal error: mid: %d, aid: %d, err: %v",
		//	frontendSess.ClientAddr(), clientRequest.MID, clientRequest.AID, err)
		return
	}
	outPacket := packet.NewFromData(dataload, nil, packet.NoneCompresser)
	outPacket.SetConnID(sid)
	outPacket.SetProtoMID(inPacket.GetProtoMID())
	outPacket.SetProtoAID(inPacket.GetProtoAID())
	outPacket.SetProtoVer(inPacket.GetProtoVer())
	outPacket.Encrypt(packet.XORCrypto)

	// zaplog.S.Debugf(
	//	"client@%s response: mid: %d, aid: %d, out: %v",
	//	frontendSess.ClientAddr(), clientRequest.MID, clientRequest.AID, outPacket,
	// )

	_, err = frontendSess.Write(outPacket)
	if err != nil {
		// TODO: log
		//zaplog.S.Errorf(
		//	"write agent@%s response: mid: %d, aid: %d, err: %v",
		//	frontendSess.ClientAddr(), clientRequest.MID,
		//	clientRequest.AID, zeroutil.ParseNetError(err))
	}
}
