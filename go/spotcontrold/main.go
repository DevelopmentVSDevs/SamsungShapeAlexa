package main

import (
    "log"
    "net/http"
    "flag"
    "fmt"
    "time"
    "errors"
    "github.com/badfortrains/spotcontrol"
    "strconv"
    "sync"
    "strings"
    "os"
    "github.com/bakins/net-http-recover"
    "github.com/gorilla/handlers"
    "github.com/gorilla/mux"
    "github.com/gorilla/rpc"
    "github.com/gorilla/rpc/json"
    "github.com/justinas/alice"

)

var username *string
var password *string
var devicename *string
var deviceid string = ""

var mutex = &sync.Mutex{}

var sController *spotcontrol.SpircController


type Irequest struct {
        I string
}

type Spotify int


func (t *Spotify) Continue(r *http.Request, args *Irequest, reply *int) error {
        if (deviceid=="") {
            return errors.New("Play device not identified")
        }

        err:= sController.SendPlay(deviceid)
        fmt.Println(err)

        return nil
}

func (t *Spotify) Pause(r *http.Request, args *Irequest, reply *int) error {
        if (deviceid=="") {
            return errors.New("Play device not identified")
        }

        err:= sController.SendPause(deviceid)
        fmt.Println(err)

        return nil
}

func (t *Spotify) PlayPlaylist(r *http.Request, args *Irequest, reply *int) error {
        if (deviceid=="") {
            return errors.New("Play device not identified")
        }
        // Clean up the playlist
        args.I = strings.Replace(args.I, ":", "/", -1)

        playlist, err := sController.GetPlaylist(args.I)
            if err != nil || playlist.Contents == nil {
                    fmt.Println("Playlist not found")
                    return errors.New("Playlist not found")
            }

        items := playlist.Contents.Items
        var ids []string
        for i := 0; i < len(items); i++ {
                id := strings.TrimPrefix(items[i].GetUri(), "spotify:track:")
                ids = append(ids, id)
        }

        sController.LoadTrack(deviceid, ids )
        sController.SendPlay(deviceid)
        return nil
}

func (t *Spotify) SetVolume(r *http.Request, args *Irequest, reply *int) error {
        v,err:=(strconv.ParseUint(args.I,10,32))
        if (deviceid!="") {
            sController.SendVolume(deviceid,uint32(v))

            fmt.Println(err)
        }

        return nil
}

func (t *Spotify) Playlist(r *http.Request, args *Irequest, names *[][]string) error {
        var err error

        playlist, _ := sController.GetRootPlaylist()
        if err != nil || playlist.Contents == nil {

                return errors.New("Error getting root list")
        }
        items := playlist.Contents.Items
        for i := 0; i < len(items); i++ {
                id := strings.TrimPrefix(items[i].GetUri(), "spotify:")
                id = strings.Replace(id, ":", "/", -1)
                list, _ := sController.GetPlaylist(id)
                //storage:=new(plist)
                //storage.title=list.Attributes.GetName()
                //storage.id=id
                //fmt.Println(*storage)
                tmp:=make([]string,2);
                tmp[0]=list.Attributes.GetName()
                tmp[1]=id
                *names = append(*names, tmp)
                fmt.Printf("PLAYLIST:%s:%s\n",list.Attributes.GetName(), id)
        }


	return nil
}



func startServer() {


	r := mux.NewRouter()

	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")


	spotify := new(Spotify)
	s.RegisterService(spotify, "")


	chain := alice.New(
		func(h http.Handler) http.Handler {
			return handlers.CombinedLoggingHandler(os.Stdout, h)
		},
		handlers.CompressHandler,
		func(h http.Handler) http.Handler {
			return recovery.Handler(os.Stderr, h, true)
		})

	// httprof and expvar endpoints
	r.PathPrefix("/debug/").Handler(chain.Then(http.DefaultServeMux))

	r.Handle("/rpc", chain.Then(s))
        log.Fatal(http.ListenAndServe(":1234", r))
}

func t_setup_devices() {

        wait_time:=time.Second * 30

        start := time.Now().AddDate(0, 0, -1)
	for {
            elapsed := time.Since(start)


            if (deviceid=="") {
                // Wait 30 seconds
                wait_time=time.Second * 30
            } else {
                // Wait 10 Minutes
                wait_time=time.Second *60*10
            }

            if (elapsed>wait_time) {

                // Lets assume this isn't re-entrant
                mutex.Lock()
                fmt.Printf("Identify Devices (LOCK)\n")
                devices := sController.ListDevices()
                for i, d := range devices {
                        fmt.Printf("IDENT %v) %v %v \n", i, d.Name, d.Ident)
			if (d.Name==*devicename) {
				fmt.Printf("FOUND %v) %v %v \n", i, d.Name, d.Ident)

                                deviceid=d.Ident
                                fmt.Println(deviceid)
                        }

                }
                fmt.Printf("Identify Devices (UNLOCK)\n")
                mutex.Unlock()
                fmt.Println(elapsed)
                start = time.Now()
                }

                time.Sleep(time.Second * 5)
            }

}



func main() {

    username = flag.String("username", "", "spotify username")
    password = flag.String("password", "", "spotify password")
    devicename = flag.String("dev", "", "name of device")
    flag.Parse()

    if *username == "" || *password == "" || *devicename == "" {
        fmt.Println("need to supply a username, password and a default device")
        return;
    }

    var err error
    sController, err = spotcontrol.Login(*username, *password, "SpotControl")

    if err != nil {
	    fmt.Println("Error logging in: ", err)
	    return
    }

    // Start the Play Device Identifier
    go t_setup_devices()
    // Start the JSON-RPC Server
    startServer()

}