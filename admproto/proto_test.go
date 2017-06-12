package admproto

import "testing"

func TestReaderWriter(t *testing.T) {
	var (
		buf []byte
		r   Reader
		w   Writer
	)

	buf = w.Begin(buf, "testapp", []byte("ins-id"))
	buf = w.Append(buf, "hello", 0)
	buf = w.Append(buf, "hello.world", 1)
	buf = w.Append(buf, "hello.foobar", 2)
	buf = w.Append(buf, "hello.foobaz", 3)

	t.Logf("%x", buf)

	buf, application, ins_id := r.Begin(buf)
	if string(application) != "testapp" || string(ins_id) != "ins-id" {
		t.Fatal("failed on begin")
	}

	buf, key, value := r.Next(buf)
	if string(key) != "hello" || value != 0 {
		t.Fatal("failed")
	}

	buf, key, value = r.Next(buf)
	if string(key) != "hello.world" || value != 1 {
		t.Fatal("failed")
	}

	buf, key, value = r.Next(buf)
	if string(key) != "hello.foobar" || value != 2 {
		t.Fatal("failed")
	}

	buf, key, value = r.Next(buf)
	if string(key) != "hello.foobaz" || value != 3 {
		t.Fatal("failed")
	}

	if len(buf) != 0 {
		t.Fatal("failed")
	}
}
