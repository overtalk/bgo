package tunnel

import (
	"github.com/overtalk/bgo/pkg/service/packet"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func handleBackendResponse(sess *BackendSession) {
	defer func() {
		if err := recover(); err != nil {
			//TODO: add log
			//zaplog.S.Error(err)
			//zaplog.S.Error(zap.Stack("").String)
		}
		defaultBackendSessionMgr.DelSession(sess.GetID())
		sess.Close()
	}()

	// TODO: add log
	//zaplog.S.Infof("connected: agent@%s ---> backend-%d@%s",
	//	sess.conn.LocalAddr(), sess.GetID(), sess.ClientAddr())

	// it's a long session
	sess.Ping()

	// FIXME: all requests must be handled after breaking the for loop
	for {
		inRequest, err := sess.ReadRequest()
		if err == nil {
			go forwardToFrontend(sess, inRequest)
		} else {
			inRequest.Free()
			if e := errors.Cause(err); !zeroutil.IsNetTimeout(e) {
				// TODO: log
				//zaplog.S.Errorf(
				//	"backend-%d@%s: %v", sess.GetID(), sess.ClientAddr(), err)
				break
			}
		}
	}
}

// forwardToFrontend forward the backend server's response to the frontend client
func forwardToFrontend(sess *BackendSession, req *BackendRequest) {
	defer func() {
		if err := recover(); err != nil {
			// TODO: log
			//zaplog.S.Error(err)
			//zaplog.S.Error(zap.Stack("").String)
		}
		req.Free()
	}()

	inPacket := req.GetPacket()
	// find the connected frontend session
	connID := inPacket.GetConnID()
	frontendSess := sess.GetFrontendSession(connID)
	if frontendSess == nil {
		//zaplog.S.Errorf("client-%d: not found", connID)
		// TODO: log
		return
	}
	if frontendSess.IsClosed() {
		// TODO: log
		//zaplog.S.Errorf("client-%d@%s: closed",
		//	connID, frontendSess.ClientAddr())
		return
	}
	// reset the server id
	inPacket.SetConnID(sess.GetID())

	// TODO: log
	//zaplog.S.Debugf("client-%d@%s: response(%v)", connID,
	//	frontendSess.ClientAddr(), inPacket)

	// encrypt the packet
	inPacket.Encrypt(packet.XORCrypto)
	// write to frontend buffer
	_, err := frontendSess.Write(inPacket)
	if err == nil {
		// TODO: log
		//zaplog.S.Debugf("client-%d@%s: response(%d bytes) done",
		//	frontendSess.GetID(), frontendSess.ClientAddr(), len(inPacket))
	} else {
		// TODO: log
		//zaplog.S.Errorf("client-%d@%s: %v", frontendSess.GetID(), frontendSess.ClientAddr(), err)
	}
	frontendSess.DoneResponse()
}
