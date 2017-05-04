import urllib2
import json
import time


def rpc_call(url, method, args):
	data = json.dumps({
	    'id': 1,
	    'method': method,
	    'params': [args]
	}).encode()
	req = urllib2.Request(url, 
		data, 
		{'Content-Type': 'application/json'})
	f = urllib2.urlopen(req)
	response = f.read()
	print response
	return json.loads(response)

url = 'http://localhost:1234/rpc'
print rpc_call(url, "Spotify.Playlist", {})
print rpc_call(url, "Spotify.Pause", {})
time.sleep(5)
print rpc_call(url, "Spotify.Continue", {})
args = {'I': 'user/spotify/playlist/xyz'}
print rpc_call(url, "Spotify.PlayPlaylist",args) 
