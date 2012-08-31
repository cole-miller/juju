package juju_test

import (
	"io/ioutil"
	. "launchpad.net/gocheck"
	"launchpad.net/juju-core/environs"
	"launchpad.net/juju-core/environs/dummy"
	"launchpad.net/juju-core/juju"
	"launchpad.net/juju-core/state"
	coretesting "launchpad.net/juju-core/testing"
	"os"
	"path/filepath"
	stdtesting "testing"
)

func Test(t *stdtesting.T) {
	coretesting.ZkTestPackage(t)
}

type ConnSuite struct {
	coretesting.LoggingSuite
}

var _ = Suite(&ConnSuite{})

func (cs *ConnSuite) TearDownTest(c *C) {
	dummy.Reset()
	cs.LoggingSuite.TearDownTest(c)
}

func (*ConnSuite) TestNewConn(c *C) {
	home := c.MkDir()
	defer os.Setenv("HOME", os.Getenv("HOME"))
	os.Setenv("HOME", home)
	conn, err := juju.NewConn("")
	c.Assert(conn, IsNil)
	c.Assert(err, ErrorMatches, ".*: no such file or directory")

	if err := os.Mkdir(filepath.Join(home, ".juju"), 0755); err != nil {
		c.Log("Could not create directory structure")
		c.Fail()
	}
	envs := filepath.Join(home, ".juju", "environments.yaml")
	err = ioutil.WriteFile(envs, []byte(`
default:
    erewhemos
environments:
    erewhemos:
        type: dummy
        zookeeper: true
        authorized-keys: i-am-a-key
`), 0644)
	if err != nil {
		c.Log("Could not create environments.yaml")
		c.Fail()
	}

	// Just run through a few operations on the dummy provider and verify that
	// they behave as expected.
	conn, err = juju.NewConn("")
	c.Assert(err, IsNil)
	defer conn.Close()
	st, err := conn.State()
	c.Assert(st, IsNil)
	c.Assert(err, ErrorMatches, "dummy environment not bootstrapped")
	err = conn.Bootstrap(false)
	c.Assert(err, IsNil)
	st, err = conn.State()
	c.Check(err, IsNil)
	c.Check(st, NotNil)
	err = conn.Destroy()
	c.Assert(err, IsNil)

	// Close the conn (thereby closing its state) a couple of times to
	// verify that multiple closes are safe.
	c.Assert(conn.Close(), IsNil)
	c.Assert(conn.Close(), IsNil)
}

func (*ConnSuite) TestNewConnFromAttrs(c *C) {
	attrs := map[string]interface{}{
		"name":            "erewhemos",
		"type":            "dummy",
		"zookeeper":       true,
		"authorized-keys": "i-am-a-key",
	}
	conn, err := juju.NewConnFromAttrs(attrs)
	c.Assert(err, IsNil)
	defer conn.Close()
	st, err := conn.State()
	c.Assert(st, IsNil)
	c.Assert(err, ErrorMatches, "dummy environment not bootstrapped")
}

func (*ConnSuite) TestConnStateSecretsSideEffect(c *C) {
	env, err := environs.NewFromAttrs(map[string]interface{}{
		"name":            "erewhemos",
		"type":            "dummy",
		"zookeeper":       true,
		"authorized-keys": "i-am-a-key",
	})
	c.Assert(err, IsNil)
	err = env.Bootstrap(false)
	c.Assert(err, IsNil)
	info, err := env.StateInfo()
	c.Assert(err, IsNil)
	st, err := state.Open(info)
	c.Assert(err, IsNil)

	// verify we have no secret in the environ config
	cfg, err := st.EnvironConfig()
	c.Assert(err, IsNil)
	c.Assert(cfg.UnknownAttrs()["secret"], IsNil)

	conn, err := juju.NewConnFromAttrs(map[string]interface{}{
		"name":            "erewhemos",
		"type":            "dummy",
		"zookeeper":       true,
		"authorized-keys": "i-am-a-key",
	})
	c.Assert(err, IsNil)
	defer conn.Close()
	// fetch a state connection via the conn, which will 
	// push the secrets.
	st, err = conn.State()
	c.Assert(err, IsNil)
	cfg, err = st.EnvironConfig()
	c.Assert(err, IsNil)
	c.Assert(cfg.UnknownAttrs()["secret"], Equals, "pork")
}

func (cs *ConnSuite) TestConnStateDoesNotUpdateExistingSecrets(c *C) {
	cs.TestConnStateSecretsSideEffect(c)
	conn, err := juju.NewConnFromAttrs(map[string]interface{}{
		"name":            "erewhemos",
		"type":            "dummy",
		"zookeeper":       true,
		"authorized-keys": "i-am-a-key",
		"secret":          "squirrel",
	})
	c.Assert(err, IsNil)
	defer conn.Close()
	st, err := conn.State()
	c.Assert(err, IsNil)
	// check that the secret has not changed
	cfg, err := st.EnvironConfig()
	c.Assert(err, IsNil)
	c.Assert(cfg.UnknownAttrs()["secret"], Equals, "pork")
}

func (*ConnSuite) TestValidRegexps(c *C) {
	assertService := func(s string, expect bool) {
		c.Assert(juju.ValidService.MatchString(s), Equals, expect)
		c.Assert(juju.ValidUnit.MatchString(s+"/0"), Equals, expect)
		c.Assert(juju.ValidUnit.MatchString(s+"/99"), Equals, expect)
		c.Assert(juju.ValidUnit.MatchString(s+"/-1"), Equals, false)
		c.Assert(juju.ValidUnit.MatchString(s+"/blah"), Equals, false)
	}
	assertService("", false)
	assertService("33", false)
	assertService("wordpress", true)
	assertService("w0rd-pre55", true)
	assertService("foo2", true)
	assertService("foo-2", false)
	assertService("foo-2foo", true)
}
