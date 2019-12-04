package base

/*
const char* go_baselib_build_time(void)
{
static const char* psz_build_time = "["__DATE__ " " __TIME__ "]";
    return psz_build_time;
}
*/
import "C"

var (
	BuildTime = C.GoString(C.go_baselib_build_time())
)
