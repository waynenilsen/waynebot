package connector

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// TLSIMAPClient implements IMAPClient using a raw TLS connection
// and minimal IMAP commands (LOGIN, SELECT, SEARCH UNSEEN, FETCH, STORE).
type TLSIMAPClient struct {
	conn io.ReadWriteCloser
	scan *bufio.Scanner
	tag  int
}

// NewTLSIMAPClient creates a new real IMAP client.
func NewTLSIMAPClient() *TLSIMAPClient {
	return &TLSIMAPClient{}
}

func (c *TLSIMAPClient) Connect(host string, port int, user, pass string) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 15 * time.Second}, "tcp", addr, &tls.Config{
		ServerName: host,
	})
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	c.conn = conn
	c.scan = bufio.NewScanner(conn)
	c.scan.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	c.tag = 0

	// Read greeting.
	if _, err := c.readLine(); err != nil {
		c.conn.Close()
		return fmt.Errorf("read greeting: %w", err)
	}

	// LOGIN
	if err := c.command(fmt.Sprintf("LOGIN %q %q", user, pass)); err != nil {
		c.conn.Close()
		return fmt.Errorf("login: %w", err)
	}

	// SELECT INBOX
	if err := c.command("SELECT INBOX"); err != nil {
		c.conn.Close()
		return fmt.Errorf("select inbox: %w", err)
	}

	return nil
}

func (c *TLSIMAPClient) FetchUnseen() ([]EmailMessage, error) {
	// SEARCH UNSEEN
	c.tag++
	tag := fmt.Sprintf("A%04d", c.tag)
	line := fmt.Sprintf("%s SEARCH UNSEEN\r\n", tag)
	if _, err := io.WriteString(c.conn, line); err != nil {
		return nil, err
	}

	var uids []string
	for {
		resp, err := c.readLine()
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(resp, "* SEARCH") {
			parts := strings.Fields(resp)
			if len(parts) > 2 {
				uids = parts[2:]
			}
		}
		if strings.HasPrefix(resp, tag) {
			if !strings.Contains(resp, "OK") {
				return nil, fmt.Errorf("search: %s", resp)
			}
			break
		}
	}

	if len(uids) == 0 {
		return nil, nil
	}

	var msgs []EmailMessage
	for _, uid := range uids {
		msg, err := c.fetchMessage(uid)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func (c *TLSIMAPClient) fetchMessage(seq string) (EmailMessage, error) {
	c.tag++
	tag := fmt.Sprintf("A%04d", c.tag)
	line := fmt.Sprintf("%s FETCH %s (BODY[HEADER.FIELDS (FROM SUBJECT DATE)] BODY[TEXT])\r\n", tag, seq)
	if _, err := io.WriteString(c.conn, line); err != nil {
		return EmailMessage{}, err
	}

	var em EmailMessage
	var inHeader, inBody bool
	var bodyLines []string

	for {
		resp, err := c.readLine()
		if err != nil {
			return em, err
		}

		if strings.HasPrefix(resp, tag) {
			break
		}

		if strings.Contains(resp, "BODY[HEADER.FIELDS") {
			inHeader = true
			inBody = false
			continue
		}
		if strings.Contains(resp, "BODY[TEXT]") {
			inBody = true
			inHeader = false
			continue
		}

		if inHeader {
			lower := strings.ToLower(resp)
			switch {
			case strings.HasPrefix(lower, "from:"):
				em.From = strings.TrimSpace(resp[5:])
			case strings.HasPrefix(lower, "subject:"):
				em.Subject = strings.TrimSpace(resp[8:])
			case strings.HasPrefix(lower, "date:"):
				if t, err := time.Parse(time.RFC1123Z, strings.TrimSpace(resp[5:])); err == nil {
					em.Date = t
				}
			case resp == "" || resp == ")":
				inHeader = false
			}
		} else if inBody {
			if resp == ")" || resp == "" && strings.HasSuffix(resp, ")") {
				inBody = false
			} else {
				bodyLines = append(bodyLines, resp)
			}
		}
	}

	em.Body = strings.Join(bodyLines, "\n")
	// Parse seq as UID.
	fmt.Sscanf(seq, "%d", &em.UID)
	return em, nil
}

func (c *TLSIMAPClient) MarkSeen(uids []uint32) error {
	for _, uid := range uids {
		if err := c.command(fmt.Sprintf("STORE %d +FLAGS (\\Seen)", uid)); err != nil {
			return err
		}
	}
	return nil
}

func (c *TLSIMAPClient) Close() error {
	if c.conn == nil {
		return nil
	}
	// Try LOGOUT, ignore errors.
	c.tag++
	tag := fmt.Sprintf("A%04d", c.tag)
	io.WriteString(c.conn, fmt.Sprintf("%s LOGOUT\r\n", tag))
	return c.conn.Close()
}

// command sends a tagged IMAP command and reads until the tagged OK response.
func (c *TLSIMAPClient) command(cmd string) error {
	c.tag++
	tag := fmt.Sprintf("A%04d", c.tag)
	line := fmt.Sprintf("%s %s\r\n", tag, cmd)
	if _, err := io.WriteString(c.conn, line); err != nil {
		return err
	}
	for {
		resp, err := c.readLine()
		if err != nil {
			return err
		}
		if strings.HasPrefix(resp, tag) {
			if !strings.Contains(resp, "OK") {
				return fmt.Errorf("%s", resp)
			}
			return nil
		}
	}
}

func (c *TLSIMAPClient) readLine() (string, error) {
	if !c.scan.Scan() {
		if err := c.scan.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return c.scan.Text(), nil
}
