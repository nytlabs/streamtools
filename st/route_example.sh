clear
echo "hi."
echo "this is the route example!"
echo "first, let's make a To Log block."
echo "(you dont have to enter the commands, I will do that for you)"
echo ""
echo -e "\033[0;36mcurl \"localhost:7070/create?blockType=tolog\"\033[0m"
curl "localhost:7070/create?blockType=tolog"
echo ""
echo -e "\033[1;33m[press enter to continue]\033[0m"
read
clear
echo "cool. now let's make a Route Example block"
echo ""
echo -e "\033[0;36mcurl \"localhost:7070/create?blockType=routeexample&id=MyRoute\"\033[0m"
curl "localhost:7070/create?blockType=routeexample&id=MyRoute"
echo ""
echo -e "\033[1;33m[press enter to continue]\033[0m"
read
clear
echo "now let's look at our blocks. you'll see something new."
echo ""
curl "localhost:7070"
echo ""
echo "blocks can have names now!"
echo ""
echo -e "\033[1;33m[press enter to continue]\033[0m"
read
clear
echo "now, lets attach the blocks together..."å
echo ""
echo -e "\033[0;36mcurl \"localhost:7070/connect?from=MyRoute&to=1\"\033[0m"
curl "localhost:7070/connect?from=MyRoute&to=1"
echo ""
echo -e "\033[1;33m[press enter to continue]\033[0m"
read
clear
echo ""
echo "OK."
echo "if you look at the log, you should see what looks like a ticker"
echo ""
echo -e "\033[1;33m[press enter to continue]\033[0m"
echo ""
read
clear
echo "now, let's see what MyRoute's rule looks like:"
echo ""
echo -e "\033[0;36mcurl \"localhost:7070/blocks/MyRoute/getRule\"\033[0m"
echo ""
curl "localhost:7070/blocks/MyRoute/getRule"
echo ""
echo ""
echo "looks like it has a period of 1 second!"
echo ""
echo -e "\033[1;33m[press enter to continue]\033[0m"
echo ""
read
clear
echo ""
echo "let's change that rule!"
echo ""
echo -e "\033[0;36mcurl --data '{\"period\":100}' \"localhost:7070/blocks/MyRoute/setRule\"\033[0m"
curl --data '{"period":100}' "localhost:7070/blocks/MyRoute/setRule"
echo ""
sleep 1
echo ""
echo "whoaaa look at that"
echo ""
echo -e "\033[1;33m[press enter to continue]\033[0m"
echo ""
read
clear
echo "ok, but rules aren't the only thing you can do with routes."
echo ""
sleep 2
echo -e "\033[0;36mlook over at log!\033[0m"
echo ""
echo ""
echo ""
sleep 3
curl --data '{"period":500000}' "localhost:7070/blocks/MyRoute/setRule"
sleep 1
curl --data '{"message":"I set the period to 500 seconds so we can chat over here now."}' "localhost:7070/blocks/MyRoute/writeMsg"
sleep 2
curl --data '{"message":"You can set abritrary handlers"}' "localhost:7070/blocks/MyRoute/writeMsg"
sleep 2
curl --data '{"message":"for anything that you want"}' "localhost:7070/blocks/MyRoute/writeMsg"
sleep 2
curl --data '{"message":"this is using the /writeMsg handler on MyRoute"}' "localhost:7070/blocks/MyRoute/writeMsg"
sleep 2
curl --data '{"message":"though it could be used for anything."}' "localhost:7070/blocks/MyRoute/writeMsg"
sleep 1
curl --data '{"message":"the end!"}' "localhost:7070/blocks/MyRoute/writeMsg"
