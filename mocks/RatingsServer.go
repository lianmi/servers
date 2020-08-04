// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"
import mock "github.com/stretchr/testify/mock"
import proto "github.com/lianmi/servers/api/proto"

// RatingsServer is an autogenerated mock type for the RatingsServer type
type RatingsServer struct {
	mock.Mock
}

// Get provides a mock function with given fields: _a0, _a1
func (_m *RatingsServer) Get(_a0 context.Context, _a1 *proto.GetRatingRequest) (*proto.Rating, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *proto.Rating
	if rf, ok := ret.Get(0).(func(context.Context, *proto.GetRatingRequest) *proto.Rating); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*proto.Rating)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *proto.GetRatingRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
