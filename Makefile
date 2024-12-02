define DESKTOPFILE
#!/usr/bin/env xdg-open
[Desktop Entry]
Terminal=false
Type=Application
Name=SugarMateReader
Exec=SugarMateReader
Icon=SugarMateReader
endef

build:
	go build -o ./SugarMateReader .

export DESKTOPFILE
install:
	/usr/local/go/bin/go build -o ./SugarMateReader .
	cp SugarMateReader /usr/local/bin
	echo "$$DESKTOPFILE" > /usr/share/applications/SugarMateReader.desktop
	chmod +x /usr/share/applications/SugarMateReader.desktop
	cp SugarMateReader.png /usr/share/icons/

uninstall:
	rm -rf /usr/local/bin/SugarMateReader
	rm -rf /usr/share/applications/SugarMateReader.desktop
	rm -rf /usr/share/icons/SugarMateReader.png