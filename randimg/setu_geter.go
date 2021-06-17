package randimg

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Yiwen-Chan/ZeroBot-Plugin/api/msgext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var (
	RANDOM_API_URL          = "https://api.pixivweb.com/anime18r.php?return=img"
	CLASSIFY_RANDOM_API_URL = "http://saki.fumiama.top:62002/dice?url=" + RANDOM_API_URL
	VOTE_API_URL            = "http://saki.fumiama.top/vote?uuid=零号&img=%s&class=%d"
	BLOCK_REQUEST           = false
	CACHE_IMG_FILE          = "/tmp/setugt"
	CACHE_URI               = "file:///" + CACHE_IMG_FILE
	last_message_id         int64
	last_dhash              string
	last_visit              = 0
)

func init() { // 插件主体
	zero.OnRegex(`^设置随机图片网址(.*)$`, zero.SuperUserPermission).SetBlock(true).SetPriority(20).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["regex_matched"].([]string)[1]
			if !strings.HasPrefix(url, "http") {
				ctx.Send("URL非法!")
			} else {
				RANDOM_API_URL = url
			}
			return
		})
	// 随机图片
	zero.OnFullMatch("随机图片").SetBlock(true).SetPriority(24).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.GroupID > 0 {
				if BLOCK_REQUEST && time.Now().Unix()-last_message_id < 30 {
					ctx.Send("请稍后再试哦")
				} else {
					last_message_id = time.Now().Unix()
					BLOCK_REQUEST = true
					if CLASSIFY_RANDOM_API_URL != "" {
						resp, err := http.Get(CLASSIFY_RANDOM_API_URL)
						if err != nil {
							ctx.Send(fmt.Sprintf("ERROR: %v", err))
						} else {
							class, err1 := strconv.Atoi(resp.Header.Get("Class"))
							dhash := resp.Header.Get("DHash")
							if err1 != nil {
								ctx.Send(fmt.Sprintf("ERROR: %v", err1))
							} else {
								defer resp.Body.Close()
								// 写入文件
								data, _ := ioutil.ReadAll(resp.Body)
								f, _ := os.OpenFile(CACHE_IMG_FILE, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
								defer f.Close()
								f.Write(data)
								if class > 4 {
									ctx.Send("太涩啦，不发了！")
									if dhash != "" {
										b14, err3 := url.QueryUnescape(dhash)
										if err3 == nil {
											ctx.Send("给你点提示哦：" + b14)
										}
									}
								} else {
									last_message_id = ctx.Send(msgext.ImageNoCache(CACHE_URI))
									last_dhash = dhash
									if class > 2 {
										ctx.Send("我好啦！")
									}
								}
							}
						}
					} else {
						ctx.Send(msgext.ImageNoCache(RANDOM_API_URL))
					}
					BLOCK_REQUEST = false
				}
			}
			return
		})
	zero.OnFullMatch("不许好").SetBlock(true).SetPriority(24).
		Handle(func(ctx *zero.Ctx) {
			if last_message_id != 0 {
				ctx.DeleteMessage(last_message_id)
				last_message_id = 0
				vote(5)
			}
		})
	zero.OnFullMatch("太涩了").SetBlock(true).SetPriority(24).
		Handle(func(ctx *zero.Ctx) {
			if last_message_id != 0 {
				ctx.DeleteMessage(last_message_id)
				last_message_id = 0
				vote(6)
			}
		})
}

func vote(class int) {
	http.Get(fmt.Sprintf(VOTE_API_URL, last_dhash, class))
}
