define get_version
$(shell cat current_version)
endef

VERSION=$(call get_version,)

install-autotag:
	wget -O autotag https://github.com/pantheon-systems/autotag/releases/download/1.1.1/Linux && sudo chmod +x autotag && sudo mv autotag /usr/local/bin/

run-rebase:
	git rebase -p master 2>/dev/null | grep "Your branch is up-to-date with 'origin/master'." || echo "\nPlease rebase your branch with master!"

run-tests:
	go test -failfast -cover ./...

build-package:
	go mod vendor

tag-version:
	autotag && git push --tags
