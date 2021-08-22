package leet

import (
	//"fmt"
	"time"

	"github.com/beevik/ntp"
)

//
//func checkNtp(server string) {
//	res, err := ntp.Query(server)
//	if nil != err {
//		_log.Error(err)
//		return
//	}
//	//fmt.Printf("%+v\n", res)
//	fmt.Printf("Time           : %+v\n", res.Time)
//	fmt.Printf("ClockOffset    : %+v ( %d )\n", res.ClockOffset, res.ClockOffset)
//	fmt.Printf("RTT            : %+v\n", res.RTT)
//	fmt.Printf("Precision      : %+v\n", res.Precision)
//	fmt.Printf("Stratum        : %d\n", res.Stratum)
//	fmt.Printf("ReferenceID    : %d\n", res.ReferenceID)
//	fmt.Printf("ReferenceTime  : %+v\n", res.ReferenceTime)
//	fmt.Printf("RootDelay      : %+v\n", res.RootDelay)
//	fmt.Printf("RootDispersion : %+v\n", res.RootDispersion)
//	fmt.Printf("RootDistance   : %+v\n", res.RootDistance)
//	fmt.Printf("Leap           : %+v\n", res.Leap)
//	fmt.Printf("MinError       : %+v\n", res.MinError)
//	fmt.Printf("KissCode       : %s\n", res.KissCode)
//	fmt.Printf("Poll           : %+v\n", res.Poll)
//}
//

func getNtpOffset(server string) (time.Duration, error) {
	res, err := ntp.Query(server)
	if nil != err {
		_log.Error().
			Err(err).
			Str("func", "getNtpOffset").
			Send()
		return time.Duration(0), err
	}
	return res.ClockOffset, nil
}
