package admproto

import "testing"

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%+v", err)
	}
}

func TestReaderWriter(t *testing.T) {
	runTest := func(t *testing.T, options Options) {
		var (
			buf []byte
			r   Reader
			w   = NewWriterWith(options)
			err error
		)

		buf, err = w.Begin(buf, "testapp", []byte("ins-id"))
		assertNoError(t, err)
		buf, err = w.Append(buf, "hello", 0)
		assertNoError(t, err)
		buf, err = w.Append(buf, "hello.world", 1)
		assertNoError(t, err)
		buf, err = w.Append(buf, "hello.foobar", 2)
		assertNoError(t, err)
		buf, err = w.Append(buf, "hello.foobaz", 3)
		assertNoError(t, err)

		t.Logf("%x", buf)

		buf, application, ins_id, err := r.Begin(buf)
		assertNoError(t, err)
		if string(application) != "testapp" || string(ins_id) != "ins-id" {
			t.Fatal("failed on begin")
		}

		buf, key, value, err := r.Next(buf)
		assertNoError(t, err)
		if string(key) != "hello" || value != 0 {
			t.Fatal("failed", string(key), value)
		}

		buf, key, value, err = r.Next(buf)
		assertNoError(t, err)
		if string(key) != "hello.world" || value != 1 {
			t.Fatal("failed", string(key), value)
		}

		buf, key, value, err = r.Next(buf)
		assertNoError(t, err)
		if string(key) != "hello.foobar" || value != 2 {
			t.Fatal("failed", string(key), value)
		}

		buf, key, value, err = r.Next(buf)
		assertNoError(t, err)
		if string(key) != "hello.foobaz" || value != 3 {
			t.Fatal("failed", string(key), value)
		}

		if len(buf) != 0 {
			t.Fatal("failed")
		}
	}

	t.Run("Float16", func(t *testing.T) {
		runTest(t, Options{FloatEncoding: Float16Encoding})
	})
	t.Run("Float32", func(t *testing.T) {
		runTest(t, Options{FloatEncoding: Float32Encoding})
	})
	t.Run("Float64", func(t *testing.T) {
		runTest(t, Options{FloatEncoding: Float64Encoding})
	})
}
