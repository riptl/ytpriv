package data

type FormatType uint8

const (
	FormatVideoOnly = 1 << iota
	FormatAudioOnly
	Format3D
	FormatHLS
	FormatDASH
	FormatHighFps
	FormatStd = 0
)

type Format struct {
	FormatID string
	Extension string
	Width uint32
	Height uint32
	VideoCodec string
	AudioCodec string
	AudioBitrate uint32
	Flags FormatType
}

// Taken from github.com/rg3/youtube-dl
// As in youtube_dl/extractor/youtube.py
var Formats = []Format{
	// Standard formats
	{ "5",  "flv",  400, 240, "h263", "mp3",  64, FormatStd },
	{ "6",  "flv",  450, 270, "h263", "mp3",  64, FormatStd },
	{ "13", "3gp",    0,   0, "mp4v", "aac",   0, FormatStd },
	{ "17", "3gp",  176, 144, "mp4v", "aac",  24, FormatStd },
	{ "18", "mp4",  640, 360, "h264", "aac",  96, FormatStd },
	{ "22", "mp4", 1280, 720, "h264", "aac", 192, FormatStd },
	{ "34", "flv",  640, 360, "h264", "aac", 128, FormatStd },
	{ "35", "flv",  854, 480, "h264", "aac", 128, FormatStd },
	// * ID 36 videos are either 320x180 (BaW_jenozKc) or 320x240 (__2ABJjxzNo), abr varies as well
	{ "36", "3gp",   320,    0, "mp4v", "aac",      0, FormatStd },
	{ "37", "mp4",  1920, 1080, "h264", "aac",    192, FormatStd },
	{ "38", "mp4",  4096, 3072, "h264", "aac",    192, FormatStd },
	{ "43", "webm",  640,  360, "vp8",  "vorbis", 128, FormatStd },
	{ "44", "webm",  854,  480, "vp8",  "vorbis", 128, FormatStd },
	{ "45", "webm", 1280,  720, "vp8",  "vorbis", 192, FormatStd },
	{ "46", "webm", 1920, 1080, "vp8",  "vorbis", 192, FormatStd },
	{ "59", "mp4",   854,  480, "h264", "aac",    128, FormatStd },
	{ "78", "mp4",   854,  480, "h264", "aac",    128, FormatStd },

	// 3D videos
	{ "82",  "mp4",  0,  360, "h264", "aac",    128, Format3D },
	{ "83",  "mp4",  0,  480, "h264", "aac",    128, Format3D },
	{ "84",  "mp4",  0,  720, "h264", "aac",    192, Format3D },
	{ "85",  "mp4",  0, 1080, "h264", "aac",    192, Format3D },
	{ "100", "webm", 0,  360, "vp8",  "vorbis", 128, Format3D },
	{ "101", "webm", 0,  480, "vp8",  "vorbis", 192, Format3D },
	{ "102", "webm", 0,  720, "vp8",  "vorbis", 192, Format3D },

	// Apple HTTP Live Streaming
	{ "91",  "mp4", 0,  144, "h264", "aac",  48, FormatHLS },
	{ "92",  "mp4", 0,  240, "h264", "aac",  48, FormatHLS },
	{ "93",  "mp4", 0,  360, "h264", "aac", 128, FormatHLS },
	{ "94",  "mp4", 0,  480, "h264", "aac", 128, FormatHLS },
	{ "95",  "mp4", 0,  720, "h264", "aac", 256, FormatHLS },
	{ "96",  "mp4", 0, 1080, "h264", "aac", 256, FormatHLS },
	{ "132", "mp4", 0,  240, "h264", "aac",  48, FormatHLS },
	{ "151", "mp4", 0,   72, "h264", "aac",  24, FormatHLS },

	// DASH mp4 video
	{ "133", "mp4", 0,  240, "h264", "", 0, FormatDASH | FormatVideoOnly },
	{ "134", "mp4", 0,  360, "h264", "", 0, FormatDASH | FormatVideoOnly },
	{ "135", "mp4", 0,  480, "h264", "", 0, FormatDASH | FormatVideoOnly },
	{ "136", "mp4", 0,  720, "h264", "", 0, FormatDASH | FormatVideoOnly },
	{ "137", "mp4", 0, 1080, "h264", "", 0, FormatDASH | FormatVideoOnly },
	{ "138", "mp4", 0,    0, "h264", "", 0, FormatDASH | FormatVideoOnly }, // Height can vary (https://github.com/rg3/youtube-dl/issues/4559)
	{ "160", "mp4", 0,  144, "h264", "", 0, FormatDASH | FormatVideoOnly },
	{ "212", "mp4", 0,  480, "h264", "", 0, FormatDASH | FormatVideoOnly },
	{ "264", "mp4", 0, 1440, "h264", "", 0, FormatDASH | FormatVideoOnly },
	{ "298", "mp4", 0,  720, "h264", "", 0, FormatDASH | FormatVideoOnly | FormatHighFps },
	{ "299", "mp4", 0, 1080, "h264", "", 0, FormatDASH | FormatVideoOnly | FormatHighFps },
	{ "266", "mp4", 0, 2160, "h264", "", 0, FormatDASH | FormatVideoOnly },

	// DASH mp4 audio
	{ "139", "m4a", 0, 0, "", "aac",   48, FormatDASH |  FormatAudioOnly },
	{ "140", "m4a", 0, 0, "", "aac",  128, FormatDASH |  FormatAudioOnly },
	{ "141", "m4a", 0, 0, "", "aac",  256, FormatDASH |  FormatAudioOnly },
	{ "256", "m4a", 0, 0, "", "aac",    0, FormatDASH |  FormatAudioOnly },
	{ "258", "m4a", 0, 0, "", "aac",    0, FormatDASH |  FormatAudioOnly },
	{ "325", "m4a", 0, 0, "", "dtse",   0, FormatDASH |  FormatAudioOnly },
	{ "328", "m4a", 0, 0, "", "ec-3",   0, FormatDASH |  FormatAudioOnly },

	// DASH webm
	{ "167", "webm",  640,  360, "vp8", "", 0, FormatDASH | FormatVideoOnly },
	{ "168", "webm",  854,  480, "vp8", "", 0, FormatDASH | FormatVideoOnly },
	{ "169", "webm", 1280,  720, "vp8", "", 0, FormatDASH | FormatVideoOnly },
	{ "170", "webm", 1920, 1080, "vp8", "", 0, FormatDASH | FormatVideoOnly },
	{ "218", "webm",  854,  480, "vp8", "", 0, FormatDASH | FormatVideoOnly },
	{ "219", "webm",  854,  480, "vp8", "", 0, FormatDASH | FormatVideoOnly },
	{ "278", "webm",    0,  144, "vp9", "", 0, FormatDASH | FormatVideoOnly },
	{ "242", "webm",    0,  240, "vp9", "", 0, FormatDASH | FormatVideoOnly },
	{ "243", "webm",    0,  360, "vp9", "", 0, FormatDASH | FormatVideoOnly },
	{ "244", "webm",    0,  480, "vp9", "", 0, FormatDASH | FormatVideoOnly },
	{ "245", "webm",    0,  480, "vp9", "", 0, FormatDASH | FormatVideoOnly },
	{ "246", "webm",    0,  480, "vp9", "", 0, FormatDASH | FormatVideoOnly },
	{ "247", "webm",    0,  720, "vp9", "", 0, FormatDASH | FormatVideoOnly },
	{ "248", "webm",    0, 1080, "vp9", "", 0, FormatDASH | FormatVideoOnly },
	{ "271", "webm",    0, 1440, "vp9", "", 0, FormatDASH | FormatVideoOnly },
	// * ID 272 videos are either 3840x2160 (e.g. RtoitU2A-3E) or 7680x4320 (sLprVF6d7Ug)
	{ "272", "webm",    0, 2160, "vp9", "", 0, FormatDASH | FormatVideoOnly },
	{ "302", "webm",    0,  720, "vp9", "", 0, FormatDASH | FormatVideoOnly | FormatHighFps },
	{ "303", "webm",    0, 1080, "vp9", "", 0, FormatDASH | FormatVideoOnly | FormatHighFps },
	{ "308", "webm",    0, 1440, "vp9", "", 0, FormatDASH | FormatVideoOnly | FormatHighFps },
	{ "313", "webm",    0, 2160, "vp9", "", 0, FormatDASH | FormatVideoOnly },
	{ "315", "webm",    0, 2160, "vp9", "", 0, FormatDASH | FormatVideoOnly | FormatHighFps },

	// DASH webm audio
	{ "171", "webm", 0, 0, "", "vorbis", 128, FormatDASH | FormatAudioOnly },
	{ "172", "webm", 0, 0, "", "vorbis", 256, FormatDASH | FormatAudioOnly },

	// DASH webm opus audio
	{ "249", "webm", 0, 0, "", "opus",  50, FormatDASH | FormatAudioOnly },
	{ "250", "webm", 0, 0, "", "opus",  70, FormatDASH | FormatAudioOnly },
	{ "251", "webm", 0, 0, "", "opus", 160, FormatDASH | FormatAudioOnly },
}
