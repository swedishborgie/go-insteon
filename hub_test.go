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
	cfg := insteon.ModemConfigurationAutoLink | insteon.ModemConfigurationMonitor | insteon.ModemConfigurationAutoLED | insteon.ModemConfigurationDeadMan

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

func (s *HubTestSuite) TestNotReady() {
	s.mock.Expect(
		[]byte{0x02, 0x73},
		[]byte{0x15},
	)

	_, err := s.hub.GetModemConfig(s.mock.ctx)
	s.Require().ErrorIs(err, insteon.ErrNotReady)
}

func TestHubSuite(t *testing.T) {
	suite.Run(t, &HubTestSuite{})
}
