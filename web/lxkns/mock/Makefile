.PHONY: mockdata

mockdata:
	node getmockdata.ts "/tmp/$$$$-mockdata.json" && mv -f "/tmp/$$$$-mockdata.json" ../src/models/lxkns/mock/mockdata.json
	echo "done moving to src/models/lxkns/mock/mockdata.json"
