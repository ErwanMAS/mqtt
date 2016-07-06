package packet

import "testing"

// TODO: Test incorrect will QOS
func TestIncrementalConnectFlag(t *testing.T) {
	c := Connect{}
	assertConnectFlagValue(t, "incorrect connect flag: default connect flag should be empty (%d)", c.connectFlag(), 0)

	c.cleanSession = true
	assertConnectFlagValue(t, "incorrect connect flag: cleanSession is not true (%d)", c.connectFlag(), 2)

	c.SetWill("topic/a", "Disconnected", 0)
	assertConnectFlagValue(t, "incorrect connect flag: willFlag is not true (%d)", c.connectFlag(), 6)

	c.SetWill("topic/a", "Disconnected", 1)
	assertConnectFlagValue(t, "incorrect connect flag: willQOS is not properly set (%d)", c.connectFlag(), 14)

	c.SetWill("topic/a", "Disconnected", 2)
	assertConnectFlagValue(t, "incorrect connect flag: willQOS is not properly set (%d)", c.connectFlag(), 22)

	c.willRetain = true
	assertConnectFlagValue(t, "incorrect connect flag: willRetain is not properly set (%d)", c.connectFlag(), 54)

	c.username = "User1"
	assertConnectFlagValue(t, "incorrect connect flag: usernameFlag is not properly set (%d)", c.connectFlag(), 118)

	c.password = "Password"
	assertConnectFlagValue(t, "incorrect connect flag: passwordFlag is not properly set (%d)", c.connectFlag(), 246)
}

func TestConnectDecode(t *testing.T) {
	connect := NewConnect()
	connect.cleanSession = true
	connect.willFlag = true
	connect.willQOS = 1
	connect.willRetain = true
	connect.keepalive = 42
	connect.ClientID = "TestClientID"
	connect.willTopic = "test/will"
	connect.willMessage = "test message"
	connect.username = "testuser"
	connect.password = "testpass"

	buf := connect.Marshall()
	if packet, err := Read(&buf); err != nil {
		t.Errorf("cannot decode connect packet: %q", err)
	} else {
		switch p := packet.(type) {
		case *Connect:
			if *p != *connect {
				t.Errorf("unmarshalled connect does not match original (%+v) = %+v", p, connect)
			}
		}
	}
}

// Helpers

func assertConnectFlagValue(t *testing.T, message string, flag int, expected int) {
	if flag != expected {
		t.Errorf(message, flag)
	}
}
