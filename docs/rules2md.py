import fileinput
import time
import json
import requests

library = requests.get("http://localhost:7070/library").json()

for i,block in enumerate(library):

    b = {
        "Id": str(i), 
        "Type":block        
    }
    print "* %s"%block
    if "rule" not in library[block]["QueryRoutes"]:
        requests.delete("http://localhost:7070/blocks/%s"%i)
        continue
    #print b
    requests.post("http://localhost:7070/blocks", data=json.dumps(b))
    d = requests.get("http://localhost:7070/blocks/%s/rule"%i).json()
    #print d

    print "    * Rules:"
    for k in d:
        if d[k] or d[k] == 0:
            print "        * `%s`: (`%s`)"%(k,d[k]) 
        else:
            print "        * `%s`: "%k 
    time.sleep(0.1)
    requests.delete("http://localhost:7070/blocks/%s"%i)
    time.sleep(0.1)

