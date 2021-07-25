package insteon_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/suite"
	"github.com/swedishborgie/go-insteon"
)

type HubTestSuite struct {
	suite.Suite
	mock *InsteonHubMock
	hub  insteon.Hub
}

func (s *HubTestSuite) SetupTest() {
	s.hub, s.mock = newMock()
}

func (s *HubTestSuite) TestGetStatus() {
	s.mock.Expect(
		[]byte{0x02, 0x60},
		[]byte{0x02, 0x60, 0x01, 0x02, 0x03, 0x03, 0x37, 0x9c, 0x06},
	)

	info, err := s.hub.GetInfo(s.mock.ctx)
	s.Require().NoError(err)
	s.Require().Equal(insteon.Address{0x01, 0x02, 0x03}, info.Address)
	s.Require().Equal(insteon.CategoryNetworkBridge, info.Category)
	s.Require().Equal(insteon.SubCategory(0x37), info.SubCategory)
	s.Require().Equal(uint8(0x9c), info.FirmwareVersion)
}

func (s *HubTestSuite) TestGetModemConfig() {
	s.mock.Expect(
		[]byte{0x02, 0x73},
		[]byte{0x02, 0x73, 0x48, 0x03, 0x00, 0x06},
	)

	cfg, err := s.hub.GetModemConfig(s.mock.ctx)
	s.Require().NoError(err)
	s.Require().Equal(true, cfg.AutoLink())
	s.Require().Equal(true, cfg.Monitor())
	s.Require().Equal(true, cfg.AutoLED())
	s.Require().Equal(true, cfg.DeadMan())
}

func (s *HubTestSuite) TestSetModemConfig() {
	cfg := insteon.ModemConfigurationAutoLink |
		insteon.ModemConfigurationMonitor |
		insteon.ModemConfigurationAutoLED |
		insteon.ModemConfigurationDeadMan

	s.mock.Expect(
		[]byte{0x02, 0x6B, byte(cfg)},
		[]byte{0x02, 0x6B, byte(cfg), 0x06},
	)

	err := s.hub.SetModemConfig(s.mock.ctx, cfg)
	s.Require().NoError(err)
}

func (s *HubTestSuite) TestGetAllLink() {
	s.mock.Expect(
		[]byte{0x02, 0x69},
		[]byte{0x02, 0x69, 0x06, 0x02, 0x57, 0xc0, 0x00, 0x01, 0x02, 0x03, 0x02, 0x08, 0x42},
		[]byte{0x02, 0x6a},
		[]byte{0x02, 0x6a, 0x06, 0x02, 0x57, 0xc0, 0x01, 0x01, 0x02, 0x03, 0x02, 0x08, 0x42},
		[]byte{0x02, 0x6a},
		[]byte{0x02, 0x6a, 0x06, 0x02, 0x57, 0xc0, 0x00, 0x01, 0x02, 0x04, 0x02, 0x1a, 0x41},
		[]byte{0x02, 0x6a},
		[]byte{0x02, 0x6a, 0x06, 0x02, 0x57, 0xc0, 0xfe, 0x01, 0x02, 0x05, 0x03, 0x00, 0x00},
		[]byte{0x02, 0x6a},
		[]byte{0x02, 0x6a, 0x06, 0x02, 0x57, 0xc0, 0x00, 0x01, 0x02, 0x06, 0x00, 0x00, 0x00},
		[]byte{0x02, 0x6a},
		[]byte{0x02, 0x6a, 0x15},
	)

	db, err := s.hub.GetAllLinkDatabase(s.mock.ctx)
	s.Require().NoError(err)
	s.Require().Len(db, 5)

	expected := []*insteon.AllLinkRecord{
		{
			Flags:   insteon.AllLinkRecordFlagsInUse | insteon.AllLinkRecordFlagsContoller,
			Group:   0,
			Address: insteon.Address{0x01, 0x02, 0x03},
			Data:    [3]uint8{0x02, 0x08, 0x42},
		},
		{
			Flags:   insteon.AllLinkRecordFlagsInUse | insteon.AllLinkRecordFlagsContoller,
			Group:   1,
			Address: insteon.Address{0x01, 0x02, 0x03},
			Data:    [3]uint8{0x02, 0x08, 0x42},
		},
		{
			Flags:   insteon.AllLinkRecordFlagsInUse | insteon.AllLinkRecordFlagsContoller,
			Group:   0,
			Address: insteon.Address{0x01, 0x02, 0x04},
			Data:    [3]uint8{0x02, 0x1a, 0x41},
		},
		{
			Flags:   insteon.AllLinkRecordFlagsInUse | insteon.AllLinkRecordFlagsContoller,
			Group:   254,
			Address: insteon.Address{0x01, 0x02, 0x05},
			Data:    [3]uint8{0x03, 0x00, 0x00},
		},
		{
			Flags:   insteon.AllLinkRecordFlagsInUse | insteon.AllLinkRecordFlagsContoller,
			Group:   0,
			Address: insteon.Address{0x01, 0x02, 0x06},
			Data:    [3]uint8{0x00, 0x00, 0x00},
		},
	}

	s.Require().True(cmp.Equal(expected, db), cmp.Diff(expected, db))
}

func (s *HubTestSuite) TestStartAllLink() {
	s.mock.Expect(
		[]byte{0x02, 0x64, byte(insteon.LinkCodeAuto), 1},
		[]byte{0x02, 0x64, byte(insteon.LinkCodeAuto), 1, 0x06},
		nil,
		[]byte{0x02, 0x53, 0x00, 0x01, 0x01, 0x02, 0x03, 0x01, 0x02, 0x12},
	)

	done, err := s.hub.StartAllLink(s.mock.ctx, insteon.LinkCodeAuto, 1)
	s.Require().NoError(err)

	expect := &insteon.AllLinkCompleted{
		LinkCode:    0,
		Group:       1,
		Address:     insteon.Address{0x1, 0x2, 0x3},
		Category:    insteon.CategoryDimmableLighting,
		SubCategory: 2,
		Firmware:    0x12,
	}

	s.Require().True(cmp.Equal(expect, done), cmp.Diff(expect, done))
}

func (s *HubTestSuite) TestCancelAllLink() {
	s.mock.Expect(
		[]byte{0x02, 0x65},
		[]byte{0x02, 0x65, 0x06},
	)

	err := s.hub.CancelAllLink(s.mock.ctx)
	s.Require().NoError(err)
}

func (s *HubTestSuite) TestBeep() {
	s.mock.Expect(
		[]byte{0x02, 0x77},
		[]byte{0x02, 0x77, 0x06},
	)

	err := s.hub.Beep(s.mock.ctx)
	s.Require().NoError(err)
}

func (s *HubTestSuite) TestSleep() {
	s.mock.Expect(
		[]byte{0x02, 0x72},
		[]byte{0x02, 0x72, 0x06},
	)

	err := s.hub.Sleep(s.mock.ctx)
	s.Require().NoError(err)
}

func (s *HubTestSuite) TestReset() {
	s.mock.Expect(
		[]byte{0x02, 0x67},
		[]byte{0x02, 0x67, 0x06},
	)

	err := s.hub.Reset(s.mock.ctx)
	s.Require().NoError(err)
}

func (s *HubTestSuite) TestSetLED() {
	s.mock.Expect(
		[]byte{0x02, 0x6d},
		[]byte{0x02, 0x6d, 0x06},
		[]byte{0x02, 0x6e},
		[]byte{0x02, 0x6e, 0x06},
	)

	s.Require().NoError(s.hub.SetLED(s.mock.ctx, true))
	s.Require().NoError(s.hub.SetLED(s.mock.ctx, false))
}

func (s *HubTestSuite) TestGetLastSender() {
	s.mock.Expect(
		[]byte{0x02, 0x6c},
		[]byte{0x02, 0x6c, 0x06},
		[]byte{},
		[]byte{0x02, 0x57, 0x02, 0x01, 0x01, 0x02, 0x03, 0x01, 0x02, 0x03},
	)

	sender, err := s.hub.GetLastSender(s.mock.ctx)
	s.Require().NoError(err)

	expect := &insteon.AllLinkRecord{
		Flags:   insteon.AllLinkRecordFlagsLast,
		Group:   0x01,
		Address: insteon.Address{1, 2, 3},
		Data:    [3]byte{1, 2, 3},
	}

	s.Require().True(cmp.Equal(expect, sender), cmp.Diff(expect, sender))
}

func (s *HubTestSuite) TestSetDeviceCategory() {
	s.mock.Expect(
		[]byte{0x02, 0x66, 0x00, 0x02, 0x12},
		[]byte{0x02, 0x66, 0x00, 0x02, 0x12, 0x06},
	)
	err := s.hub.SetDeviceCategory(s.mock.ctx, insteon.CategoryGeneralController, 0x02, 0x12)
	s.Require().NoError(err)
}

func (s *HubTestSuite) TestNotReady() {
	s.mock.Expect(
		[]byte{0x02, 0x73},
		[]byte{0x15},
	)

	_, err := s.hub.GetModemConfig(s.mock.ctx)
	s.Require().ErrorIs(err, insteon.ErrNotReady)
}

func TestHubSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, &HubTestSuite{})
}
