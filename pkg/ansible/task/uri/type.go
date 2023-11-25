package uri

// Doc: https://docs.ansible.com/ansible/2.9/modules/uri_module.html

import (
	"fmt"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/log"
)

type TaskV1 struct {
	// HTTP or HTTPS URL in the form (http|https)://host.domain[:port]/path
	Url string `task:",req"`
	// The HTTP method of the request or response.
	Method string `task:",def:GET"`

	// Whether or not the URI module should follow redirects.
	Follow_redirects string `task:",def:safe,list:all none safe urllib2"`

	// The socket level timeout in seconds.
	Timeout int
	// A list of valid, numeric, HTTP status codes that signifies success of the request.
	Status_code []int

	// Path to file to be submitted to the remote server.
	//Src string
	// The body of the http request/response to the web service.
	//Body []byte
	// The serialization format of the body.
	//Body_format string `task:",def:raw,list:form-urlencoded json raw"`
	// Add custom HTTP headers to a request in the format of a YAML hash.
	//Headers ansible.OrderedMap
	// Header to identify as, generally appears in web server logs.
	//Http_agent string `task:",def:ansible-httpget"`
	// Whether or not to return the body of the response as a "content" key in the dictionary result.
	//Return_content bool
	// A username for the module to use for Digest, Basic or WSSE authentication.
	//Url_username string `task:",alias:user"`
	// A password for the module to use for Digest, Basic or WSSE authentication.
	//Url_password string `task:",alias:password"`

	// PEM formatted certificate chain file to be used for SSL client authentication.
	//Client_cert string
	// PEM formatted file that contains your private key to be used for SSL client authentication.
	//Client_key string
	// If no, SSL certificates will not be validated.
	//Validate_certs bool `task:"def:true"`

	// A filename, when it already exists, this step will not be run.
	//Creates string
	// A filename, when it does not exist, this step will not be run.
	//Removes string

	// If yes do not get a cached copy.
	//Force bool
	// Force the sending of the Basic authentication header upon initial request.
	//Force_basic_auth bool

	// The attributes the resulting file or directory should have.
	//Attributes string `task:",alias:attr"`
	// A path of where to download the file to (if desired).
	//Dest string
	// If no, the module will search for src on originating/master machine. If yes the module will use the src path on the remote/target machine.
	//Remote_src bool
	// Name of the user that should own the file/directory, as would be fed to chown.
	//Owner string
	// Name of the group that should own the file/directory, as would be fed to chown.
	//Group string
	// The permissions the resulting file or directory should have.
	//Mode string

	// Influence when to use atomic operation to prevent data corruption or inconsistent reads from the target file.
	//Unsafe_writes bool

	// Path to Unix domain socket to use for connection
	//Unix_socket string
	// If no, it will not use a proxy, even if one is defined in an environment variable on the target hosts.
	//Use_proxy bool `task:",def:true"`

	// The level part of the SELinux file context.
	//Selevel string `task:",def:s0"`
	// The role part of the SELinux file context.
	//Serole  string
	// The type part of the SELinux file context.
	//Setype  string
	// The user part of the SELinux file context.
	//Seuser  string
}

func (t *TaskV1) SetData(data ansible.OrderedMap) error {
	uri_data, ok := data.Get("uri")
	if !ok {
		return fmt.Errorf("Unable to find the 'uri' map in task data")
	}
	fmap, ok := uri_data.(ansible.OrderedMap)
	if !ok {
		return fmt.Errorf("The 'uri' is not the OrderedMap")
	}
	return ansible.TaskV1SetData(t, fmap)
}

func (t *TaskV1) GetData() (data ansible.OrderedMap) {
	fmap := ansible.TaskV1GetData(t)
	data.Set("uri", fmap)
	return data
}

func (t *TaskV1) Run(vars map[string]any) error {
	log.Error("TODO: Implement uri.Run")

	return nil
}
