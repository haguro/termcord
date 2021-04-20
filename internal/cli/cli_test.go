package cli_test

import (
	"bytes"
	"io"
	"regexp"
	"testing"
	"time"

	"github.com/haguro/termcord/internal/cli"
)

type mockFile struct {
	name    string
	content *bytes.Buffer
}

func (m *mockFile) Read(p []byte) (n int, err error) {
	return m.content.Read(p)
}

func (m *mockFile) Write(p []byte) (n int, err error) {
	return m.content.Write(p)
}

func (m *mockFile) Close() error {
	return nil
}

func mockRecorderSetup(mf *mockFile) cli.RecorderSetupFunc {
	return func(filename string, append bool) (io.ReadWriteCloser, error) {
		mf.name = filename
		if mf.content == nil || !append {
			mf.content = &bytes.Buffer{}
		}
		return mf, nil
	}
}

func TestRunPrintHelp(t *testing.T) {
	t.Parallel()
	args := []string{"termcord", "-h"}
	in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	f := mockFile{}

	r := cli.Run(
		args,
		cli.WithInput(in),
		cli.WithOutput(out),
		cli.WithErrorOutput(errOut),
		cli.WithRecorderSetupFunc(mockRecorderSetup(&f)),
	)

	if r != 0 {
		t.Errorf("expected to return 0, returned %d - error output was: %s", r, errOut.String())
	}
	if !bytes.Contains(errOut.Bytes(), []byte("termcord is")) {
		t.Errorf("help is expected to contain an intro line")
	}
	if !bytes.Contains(errOut.Bytes(), []byte("Usage:")) {
		t.Errorf("help is expected to contain a usage statement")
	}
	if !bytes.Contains(errOut.Bytes(), []byte("Options:")) {
		t.Errorf("help is expected to contain options descriptions")
	}
}

func TestRunRecordCommandOutput(t *testing.T) {
	t.Parallel()
	want := "hello, world"
	args := []string{"termcord", "echo", want}
	in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	f := mockFile{}

	r := cli.Run(
		args,
		cli.WithInput(in),
		cli.WithOutput(out),
		cli.WithErrorOutput(errOut),
		cli.WithRecorderSetupFunc(mockRecorderSetup(&f)),
	)

	if r != 0 {
		t.Errorf("expected to return 0, returned %d - error output was: %s", r, errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte(want)) {
		t.Errorf("expected output to contain %q", want)
	}
}

func TestRunPrintsRecordingPromptsByDefault(t *testing.T) {
	t.Parallel()
	args := []string{"termcord", "echo"}
	in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	f := mockFile{}

	r := cli.Run(
		args,
		cli.WithInput(in),
		cli.WithOutput(out),
		cli.WithErrorOutput(errOut),
		cli.WithRecorderSetupFunc(mockRecorderSetup(&f)),
	)

	if r != 0 {
		t.Errorf("expected to return 0, returned %d - error output was: %s", r, errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("Starting recording session")) {
		t.Errorf("expected output to contain a recording start prompt")
	}
	if !bytes.Contains(out.Bytes(), []byte("Recording session ended")) {
		t.Errorf("expected output to contain a recording end prompt")
	}
}

func TestRunDoesNotPrintRecordingPromptsWhenFlagIsPassed(t *testing.T) {
	t.Parallel()
	args := []string{"termcord", "-q", "echo"}
	in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	f := mockFile{}

	r := cli.Run(
		args,
		cli.WithInput(in),
		cli.WithOutput(out),
		cli.WithErrorOutput(errOut),
		cli.WithRecorderSetupFunc(mockRecorderSetup(&f)),
	)

	if r != 0 {
		t.Errorf("expected to return 0, returned %d - error output was: %s", r, errOut.String())
	}
	if bytes.Contains(out.Bytes(), []byte("Starting recording session")) {
		t.Errorf("expected output to not contain a recording start prompt")
	}
	if bytes.Contains(out.Bytes(), []byte("Recording session ended")) {
		t.Errorf("expected output to not contain a recording end prompt")
	}
}

func TestRunRecordWithDefaultFilename(t *testing.T) {
	t.Parallel()
	args := []string{"termcord", "echo"}
	in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	f := mockFile{}

	r := cli.Run(
		args,
		cli.WithInput(in),
		cli.WithOutput(out),
		cli.WithErrorOutput(errOut),
		cli.WithRecorderSetupFunc(mockRecorderSetup(&f)),
	)

	if r != 0 {
		t.Errorf("expected to return 0, returned %d - error output was: %s", r, errOut.String())
	}
	if f.name != cli.DefaultFileName {
		t.Errorf("expected filename to be %q, got %q", cli.DefaultFileName, f.name)
	}
}

func TestRunRecordWithCustomFilename(t *testing.T) {
	t.Parallel()
	want := "myfile"
	args := []string{"termcord", "-f", want, "echo"}
	in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	f := mockFile{}

	r := cli.Run(
		args,
		cli.WithInput(in),
		cli.WithOutput(out),
		cli.WithErrorOutput(errOut),
		cli.WithRecorderSetupFunc(mockRecorderSetup(&f)),
	)

	if r != 0 {
		t.Errorf("expected to return 0, returned %d - error output was: %s", r, errOut.String())
	}
	if f.name != want {
		t.Errorf("expected filename to be %q, got %q", want, f.name)
	}
}

func TestRunAddRecordingTimestampedPromptsToFile(t *testing.T) {
	t.Parallel()
	args := []string{"termcord", "echo"}
	in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	f := mockFile{}

	r := cli.Run(
		args,
		cli.WithInput(in),
		cli.WithOutput(out),
		cli.WithErrorOutput(errOut),
		cli.WithRecorderSetupFunc(mockRecorderSetup(&f)),
	)

	if r != 0 {
		t.Errorf("expected to return 0, returned %d - error output was: %s", r, errOut.String())
	}

	startRx := regexp.MustCompile("Recording started on (.+)")
	m := startRx.FindStringSubmatch(f.content.String())
	if len(m) < 2 {
		t.Errorf("expected file to contain a recording start prompt")
	} else if _, err := time.Parse(time.RFC1123, m[1]); err != nil {
		t.Errorf("expected file to contain recording start prompt and timestamp in RFC1123 format")
	}

	endRx := regexp.MustCompile("Recording ended on (.+)")
	m = endRx.FindStringSubmatch(f.content.String())
	if len(m) < 2 {
		t.Errorf("expected file to contain a recording start prompt")
	} else if _, err := time.Parse(time.RFC1123, m[1]); err != nil {
		t.Errorf("expected file to contain recording start timestamp in RFC1123 format")
	}

}
func TestRunTruncateRecordingFileByDefault(t *testing.T) {
	t.Parallel()
	existing := []byte("Existing file content")
	args := []string{"termcord", "echo"}
	in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	f := mockFile{
		content: bytes.NewBuffer([]byte(existing)),
	}

	r := cli.Run(
		args,
		cli.WithInput(in),
		cli.WithOutput(out),
		cli.WithErrorOutput(errOut),
		cli.WithRecorderSetupFunc(mockRecorderSetup(&f)),
	)

	if r != 0 {
		t.Errorf("expected to return 0, returned %d - error output was: %s", r, errOut.String())
	}

	if bytes.Contains(f.content.Bytes(), existing) {
		t.Errorf("expected pre-existing file content to be truncated")
	}
}

func TestRunAppendToRecordingFileWhenFlagIsPassed(t *testing.T) {
	t.Parallel()
	existing := []byte("Existing file content")
	args := []string{"termcord", "-a", "echo"}
	in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	f := mockFile{
		content: bytes.NewBuffer([]byte(existing)),
	}

	r := cli.Run(
		args,
		cli.WithInput(in),
		cli.WithOutput(out),
		cli.WithErrorOutput(errOut),
		cli.WithRecorderSetupFunc(mockRecorderSetup(&f)),
	)

	if r != 0 {
		t.Errorf("expected to return 0, returned %d - error output was: %s", r, errOut.String())
	}

	if !bytes.Contains(f.content.Bytes(), existing) {
		t.Errorf("expected pre-existing file content to be retained")
	}
}
