package vklongpoll_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/ciricc/vklongpoll"
)

func TestBuildOptions(t *testing.T) {
	t.Run("default values uses", func(t *testing.T) {
		defaultOpt := vklongpoll.NewOptions()
		builtOpt := vklongpoll.BuildOptions()
		if !reflect.DeepEqual(defaultOpt, builtOpt) {
			t.Errorf("expected %v but got %v options\n", defaultOpt, builtOpt)
		}
	})

	t.Run("correct sum all options", func(t *testing.T) {
		waitOpt := 10 * time.Minute
		versionOpt := 6
		modeOpt := vklongpoll.Attachments
		serverUpdater := vklongpoll.UniversalServerUpdater(nil, nil)

		opt := vklongpoll.BuildOptions(
			vklongpoll.WithMode(modeOpt),
			vklongpoll.WithVersion(versionOpt),
			vklongpoll.WithWait(waitOpt),
		)

		expectedOpt := vklongpoll.NewOptions()
		expectedOpt.Mode = modeOpt
		expectedOpt.Version = versionOpt
		expectedOpt.Wait = waitOpt

		opt2 := vklongpoll.BuildOptions(vklongpoll.WithServerUpdater(serverUpdater))

		if opt2.ServerUpdater == nil {
			t.Errorf("expected server updater non-nil")
		}

		if !reflect.DeepEqual(expectedOpt, opt) {
			t.Errorf("expected %v options but got %v\n", expectedOpt, opt)
		}
	})
}

func TestNewOptions(t *testing.T) {
	t.Run("uses custom default values", func(t *testing.T) {
		waitOpt := 1 * time.Second
		verionOpt := 1
		modeOpt := -1

		vklongpoll.DefaultWait = waitOpt
		vklongpoll.DefaultVersion = verionOpt
		vklongpoll.DefaultMode = vklongpoll.Mode(modeOpt)

		opt := vklongpoll.NewOptions()
		expectedOpt := &vklongpoll.VkLongPollOptions{
			Wait:    waitOpt,
			Mode:    vklongpoll.Mode(modeOpt),
			Version: verionOpt,
		}

		if !reflect.DeepEqual(expectedOpt, opt) {
			t.Errorf("expected options %v but got %v", expectedOpt, opt)
		}
	})
}

func TestModeSum(t *testing.T) {
	t.Run("zero modes", func(t *testing.T) {
		sum := vklongpoll.SumModes()
		if sum != 0 {
			t.Errorf("expected 0 but got %d", sum)
		}
	})
	t.Run("all modes", func(t *testing.T) {
		sum := vklongpoll.SumModes(
			vklongpoll.Attachments,
			vklongpoll.ExtraFields,
			vklongpoll.ReturnPts,
			vklongpoll.Extended,
		)
		expected := 106
		if sum != vklongpoll.Mode(expected) {
			t.Errorf("expected sum %d but got %d", expected, sum)
		}
	})
}
