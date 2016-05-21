package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/therecipe/qt/internal/binding/templater"
	"github.com/therecipe/qt/internal/utils"
)

var (
	appPath, appName       string
	depPath                string
	buildMode, buildTarget string
	ending                 string
)

func main() {

	args()

	switch buildMode {
	case "build", "test":
		{
			moc()
			qrc()
			build()
			predeploy()
			deploy()
			pastdeploy()
			cleanup()
		}
	}

	switch buildMode {
	case "run", "test":
		{
			run()
		}
	}
}

func args() {

	switch len(os.Args) {
	case 1:
		buildMode = "test"
		buildTarget = "desktop"
		appPath, _ = os.Getwd()

	case 2:
		buildMode = os.Args[1]
		buildTarget = "desktop"
		appPath, _ = os.Getwd()

	case 3:
		buildMode = os.Args[1]
		buildTarget = os.Args[2]
		appPath, _ = os.Getwd()

	case 4:
		buildMode = os.Args[1]
		buildTarget = os.Args[2]
		appPath = os.Args[3]
	}

	switch buildMode {
	case "build", "run", "test":
		{
			switch buildTarget {
			case "desktop", "android", "ios", "ios-simulator":
				{

				}

			default:
				{
					fmt.Println("usage:", "qtdeploy", "[ build | run | test ]", "[ desktop | android | ios | ios-simulator ]", filepath.Join("path", "to", "project"))
					os.Exit(1)
				}
			}
		}

	default:
		{
			fmt.Println("usage:", "qtdeploy", "[ build | run | test ]", "[ desktop | android | ios | ios-simulator ]", filepath.Join("path", "to", "project"))
			os.Exit(1)
		}
	}

	if !filepath.IsAbs(appPath) {
		appPath = utils.GetAbsPath(appPath)
	}
	appName = filepath.Base(appPath)

	switch buildTarget {
	case "android", "ios", "ios-simulator":
		{
			depPath = filepath.Join(appPath, "deploy", buildTarget)
		}

	case "desktop":
		{
			depPath = filepath.Join(appPath, "deploy", runtime.GOOS)
		}
	}

	switch buildMode {
	case "build", "test":
		{
			utils.RemoveAll(depPath)
			utils.MakeFolder(depPath)
		}
	}

	if runtime.GOOS == "windows" && buildTarget == "desktop" {
		ending = ".exe"
	}
}

func moc() {
	var moc = exec.Command("qtmoc")
	moc.Dir = appPath
	runCmd(moc, "qtdeploy.moc")
}

func qrc() {

	utils.MakeFolder(filepath.Join(appPath, "qml"))

	var (
		rccPath string
		qmlGo   = filepath.Join(appPath, "qrc.go")
		qmlQrc  = filepath.Join(appPath, "qrc.qrc")
		qmlCpp  = filepath.Join(appPath, "qrc.cpp")
	)

	switch buildTarget {
	case "android":
		{
			switch runtime.GOOS {
			case "darwin":
				{
					rccPath = "/usr/local/Qt5.6.0/5.6/android_armv7/bin/rcc"
				}

			case "linux":
				{

					switch runtime.GOARCH {
					case "amd64":
						{
							rccPath = "/usr/local/Qt5.6.0/5.6/android_armv7/bin/rcc"
						}

					case "386":
						{
							rccPath = "/usr/local/Qt5.6.0/5.6/android_armv7/bin/rcc"
						}
					}
				}

			case "windows":
				{
					rccPath = "C:\\Qt\\Qt5.6.0\\5.6\\android_armv7\\bin\\rcc.exe"
				}
			}
		}

	case "ios", "ios-simulator":
		{
			switch runtime.GOOS {
			case "darwin":
				{
					rccPath = "/usr/local/Qt5.6.0/5.6/ios/bin/rcc"
				}
			}
		}

	case "desktop":
		{
			switch runtime.GOOS {
			case "darwin":
				{
					rccPath = "/usr/local/Qt5.6.0/5.6/clang_64/bin/rcc"
				}

			case "linux":
				{

					switch runtime.GOARCH {
					case "amd64":
						{
							rccPath = "/usr/local/Qt5.6.0/5.6/gcc_64/bin/rcc"
						}

					case "386":
						{
							rccPath = "/usr/local/Qt5.6.0/5.6/gcc/bin/rcc"
						}
					}
				}

			case "windows":
				{
					rccPath = "C:\\Qt\\Qt5.6.0\\5.6\\mingw49_32\\bin\\rcc.exe"
				}
			}
		}
	}

	utils.Save(qmlGo, qmlHeader())

	var rcc = exec.Command(rccPath, "-project", "-o", qmlQrc)
	rcc.Dir = filepath.Join(appPath, "qml")
	runCmd(rcc, "qrc.qrc")

	utils.Save(qmlQrc, strings.Replace(utils.Load(qmlQrc), "<file>./", "<file>qml/", -1))

	rcc = exec.Command(rccPath, "-name", appName, "-o", qmlCpp, qmlQrc)
	runCmd(rcc, "qrc.cpp")
}

func qmlHeader() string {

	var hloc = func() string {
		if runtime.GOOS == "windows" {
			return "C:/Qt"
		}
		return "/usr/local"
	}()

	return fmt.Sprintf(`package main

/*
#cgo +build windows,386 LDFLAGS: -LC:/Qt/Qt5.6.0/5.6/mingw49_32/lib -lQt5Core

#cgo +build darwin,amd64 LDFLAGS: -F/usr/local/Qt5.6.0/5.6/clang_64/lib -framework QtCore

#cgo +build linux,amd64 LDFLAGS: -Wl,-rpath,/usr/local/Qt5.6.0/5.6/gcc_64/lib
#cgo +build linux,amd64 LDFLAGS: -L/usr/local/Qt5.6.0/5.6/gcc_64/lib -lQt5Core

#cgo +build linux,386 LDFLAGS: -Wl,-rpath,/usr/local/Qt5.6.0/5.6/gcc/lib
#cgo +build linux,386 LDFLAGS: -L/usr/local/Qt5.6.0/5.6/gcc/lib -lQt5Core

#cgo +build android,arm LDFLAGS: -L%v/Qt5.6.0/5.6/android_armv7/lib -lQt5Core

#cgo +build darwin,386 LDFLAGS: -headerpad_max_install_names -stdlib=libc++ -Wl,-syslibroot,/Applications/Xcode.app/Contents/Developer/Platforms/iPhoneSimulator.platform/Developer/SDKs/iPhoneSimulator9.3.sdk -mios-simulator-version-min=6.1 -arch i386
#cgo +build darwin,386 LDFLAGS: -L/usr/local/Qt5.6.0/5.6/ios/plugins/platforms -lqios_iphonesimulator -framework Foundation -framework UIKit -framework QuartzCore -framework AssetsLibrary -L/usr/local/Qt5.6.0/5.6/ios/lib -framework MobileCoreServices -framework CoreFoundation -framework CoreText -framework CoreGraphics -framework OpenGLES -lqtfreetype_iphonesimulator -framework Security -framework SystemConfiguration -framework CoreBluetooth -L/usr/local/Qt5.6.0/5.6/ios/plugins/imageformats -lqdds_iphonesimulator -lqicns_iphonesimulator -lqico_iphonesimulator -lqtga_iphonesimulator -lqtiff_iphonesimulator -lqwbmp_iphonesimulator -lqwebp_iphonesimulator -lqtharfbuzzng_iphonesimulator -lz -lqtpcre_iphonesimulator -lm -lQt5Widgets_iphonesimulator -lQt5Core_iphonesimulator -lQt5Gui_iphonesimulator -lQt5PlatformSupport_iphonesimulator

#cgo +build darwin,arm64 LDFLAGS: -headerpad_max_install_names -stdlib=libc++ -Wl,-syslibroot,/Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/Developer/SDKs/iPhoneOS9.3.sdk -miphoneos-version-min=6.1 -arch arm64
#cgo +build darwin,arm64 LDFLAGS: -L/usr/local/Qt5.6.0/5.6/ios/plugins/platforms -lqios -framework Foundation -framework UIKit -framework QuartzCore -framework AssetsLibrary -L/usr/local/Qt5.6.0/5.6/ios/lib -framework MobileCoreServices -framework CoreFoundation -framework CoreText -framework CoreGraphics -framework OpenGLES -lqtfreetype -framework Security -framework SystemConfiguration -framework CoreBluetooth -L/usr/local/Qt5.6.0/5.6/ios/plugins/imageformats -lqdds -lqicns -lqico -lqtga -lqtiff -lqwbmp -lqwebp -lqtharfbuzzng -lz -lqtpcre -lm -lQt5Widgets -lQt5Core -lQt5Gui -lQt5PlatformSupport

#cgo +build darwin,arm LDFLAGS: -headerpad_max_install_names -stdlib=libc++ -Wl,-syslibroot,/Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/Developer/SDKs/iPhoneOS9.3.sdk -miphoneos-version-min=6.1 -arch armv7
#cgo +build darwin,arm LDFLAGS: -L/usr/local/Qt5.6.0/5.6/ios/plugins/platforms -lqios -framework Foundation -framework UIKit -framework QuartzCore -framework AssetsLibrary -L/usr/local/Qt5.6.0/5.6/ios/lib -framework MobileCoreServices -framework CoreFoundation -framework CoreText -framework CoreGraphics -framework OpenGLES -lqtfreetype -framework Security -framework SystemConfiguration -framework CoreBluetooth -L/usr/local/Qt5.6.0/5.6/ios/plugins/imageformats -lqdds -lqicns -lqico -lqtga -lqtiff -lqwbmp -lqwebp -lqtharfbuzzng -lz -lqtpcre -lm -lQt5Widgets -lQt5Core -lQt5Gui -lQt5PlatformSupport
*/
import "C"`, hloc)
}

func build() {

	var (
		ldFlags, tagFlags string
		outputFile        string
		env               map[string]string
	)

	switch buildTarget {
	case "android":
		{
			ldFlags = "-ldflags=\"-s\" \"-w\""
			outputFile = filepath.Join(depPath, "libgo_base.so")

			switch runtime.GOOS {
			case "darwin", "linux":
				{
					env = map[string]string{
						"PATH":   os.Getenv("PATH"),
						"GOPATH": os.Getenv("GOPATH"),
						"GOROOT": runtime.GOROOT(),

						"GOOS":   "android",
						"GOARCH": "arm",
						"GOARM":  "7",

						"CC":           filepath.Join("/opt", "android-ndk", "toolchains", "arm-linux-androideabi-4.9", "prebuilt", runtime.GOOS+"-x86_64", "bin", "arm-linux-androideabi-gcc"),
						"CXX":          filepath.Join("/opt", "android-ndk", "toolchains", "arm-linux-androideabi-4.9", "prebuilt", runtime.GOOS+"-x86_64", "bin", "arm-linux-androideabi-g++"),
						"CGO_ENABLED":  "1",
						"CGO_CPPFLAGS": "-isystem /opt/android-ndk/platforms/android-9/arch-arm/usr/include",
						"CGO_LDFLAGS":  "--sysroot=/opt/android-ndk/platforms/android-9/arch-arm/ -llog",
					}
				}

			case "windows":
				{
					env = map[string]string{
						"PATH":   os.Getenv("PATH"),
						"GOPATH": os.Getenv("GOPATH"),
						"GOROOT": runtime.GOROOT(),

						"GOOS":   "android",
						"GOARCH": "arm",
						"GOARM":  "7",

						"CC":           "C:\\android\\android-ndk\\toolchains\\arm-linux-androideabi-4.9\\prebuilt\\windows\\bin\\arm-linux-androideabi-gcc.exe",
						"CXX":          "C:\\android\\android-ndk\\toolchains\\arm-linux-androideabi-4.9\\prebuilt\\windows\\bin\\arm-linux-androideabi-g++.exe",
						"CGO_ENABLED":  "1",
						"CGO_CPPFLAGS": "-isystem C:\\android\\android-ndk\\platforms\\android-9\\arch-arm\\usr\\include",
						"CGO_LDFLAGS":  "--sysroot=C:\\android\\android-ndk\\platforms\\android-9\\arch-arm\\ -llog",
					}
				}
			}

			utils.Save(filepath.Join(appPath, "cgo_main_wrapper.go"), "package main\nimport \"C\"\n//export go_main_wrapper\nfunc go_main_wrapper() { main() }")
		}

	case "ios", "ios-simulator":
		{
			ldFlags = "-ldflags=\"-w\""
			tagFlags = "-tags=\"ios\""
			outputFile = filepath.Join(depPath, "libgo.a")

			var (
				GOARCH = func() string {
					if buildTarget == "ios" {
						return "arm64"
					}
					return "386"
				}()

				CLANGARCH, CLANGDIR, CLANGFLAG = func() (string, string, string) {
					if buildTarget == "ios" {
						return "arm64", "iPhoneOS", "iphoneos"
					}
					return "i386", "iPhoneSimulator", "ios-simulator"
				}()
			)

			env = map[string]string{
				"PATH":   os.Getenv("PATH"),
				"GOPATH": os.Getenv("GOPATH"),
				"GOROOT": runtime.GOROOT(),

				"GOOS":   runtime.GOOS,
				"GOARCH": GOARCH,

				"CGO_ENABLED":  "1",
				"CGO_CPPFLAGS": fmt.Sprintf("-isysroot /Applications/Xcode.app/Contents/Developer/Platforms/%v.platform/Developer/SDKs/%v9.3.sdk -m%v-version-min=6.1 -arch %v", CLANGDIR, CLANGDIR, CLANGFLAG, CLANGARCH),
				"CGO_LDFLAGS":  fmt.Sprintf("-isysroot /Applications/Xcode.app/Contents/Developer/Platforms/%v.platform/Developer/SDKs/%v9.3.sdk -m%v-version-min=6.1 -arch %v", CLANGDIR, CLANGDIR, CLANGFLAG, CLANGARCH),
			}

			utils.Save(filepath.Join(appPath, "cgo_main_wrapper.go"), "package main\nimport \"C\"\n//export go_main_wrapper\nfunc go_main_wrapper() { main() }")
		}

	case "desktop":
		{
			switch runtime.GOOS {
			case "darwin":
				{
					ldFlags = "-ldflags=\"-w\" \"-r=/usr/local/Qt5.6.0/5.6/clang_64/lib\""
					outputFile = filepath.Join(depPath, fmt.Sprintf("%v.app/Contents/MacOS/%v", appName, appName))
				}

			case "linux":
				{
					ldFlags = "-ldflags=\"-s\" \"-w\""
					outputFile = filepath.Join(depPath, appName)
				}

			case "windows":
				{
					ldFlags = "-ldflags=\"-s\" \"-w\" \"-H=windowsgui\""
					outputFile = filepath.Join(depPath, appName)
					env = map[string]string{
						"PATH":   os.Getenv("PATH"),
						"GOPATH": os.Getenv("GOPATH"),
						"GOROOT": runtime.GOROOT(),

						"GOOS":   runtime.GOOS,
						"GOARCH": "386",

						"CGO_ENABLED": "1",
					}
				}
			}
		}
	}

	var cmd = exec.Command("go", "build", ldFlags, "-o", outputFile+ending)
	cmd.Dir = appPath
	if tagFlags != "" {
		cmd.Args = append(cmd.Args, tagFlags)
	}

	if buildTarget != "desktop" || runtime.GOOS == "windows" {
		if buildTarget == "android" {
			cmd.Args = append(cmd.Args, "-buildmode", "c-shared")
		}
		if buildTarget == "ios" || buildTarget == "ios-simulator" {
			cmd.Args = append(cmd.Args, "-buildmode", "c-archive")
		}
		for key, value := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%v=%v", key, value))
		}
	}
	runCmd(cmd, "build_1")

	//armv7
	/*
		if buildTarget == "ios" {
			var cmdiOS = exec.Command("go", "build", ldFlags, "-o", strings.Replace(outputFile, "libgo.a", "libgo_armv7.a", -1))
			cmdiOS.Dir = appPath
			if tagFlags != "" {
				cmdiOS.Args = append(cmdiOS.Args, tagFlags)
			}
			cmdiOS.Args = append(cmdiOS.Args, "-buildmode", "c-archive")
			var tmp = strings.Replace(strings.Join(cmd.Env, "|"), "-arch arm64", "-arch armv7", -1)
			tmp = strings.Replace(tmp, "arm64", "arm", -1)
			cmdiOS.Env = append(strings.Split(tmp, "|"), "GOARM=7")
			runCmd(cmdiOS, "build_2")
		}
	*/
}

func predeploy() {

	var copyCmd = func() string {
		if runtime.GOOS == "windows" {
			return "xcopy"
		}
		return "cp"
	}()

	switch buildTarget {
	case "android":
		{
			utils.MakeFolder(filepath.Join(appPath, "android"))

			for _, dir := range []string{"drawable-hdpi", "drawable-ldpi", "drawable-mdpi"} {
				utils.MakeFolder(filepath.Join(depPath, "build", "res", dir))
				runCmdOptional(exec.Command(copyCmd, filepath.Join(appPath, "android", "icon.png"), filepath.Join(depPath, "build", "res", dir, "icon.png")), "predeploy.cpicon")
			}

			//add c_main_wrappers
			utils.Save(filepath.Join(depPath, "c_main_wrapper.cpp"), "#include \"libgo_base.h\"\nint main(int argc, char *argv[]) { go_main_wrapper(); }")

			var libPath = filepath.Join(depPath, "build", "libs", "armeabi-v7a")
			utils.MakeFolder(libPath)

			var (
				qtPrefix      string
				androidPrefix string
				ndkhost       string
				compiler      string
			)

			switch runtime.GOOS {
			case "darwin", "linux":
				{
					qtPrefix = "/usr/local"
					androidPrefix = "/opt"
					ndkhost = runtime.GOOS + "-x86_64"
					compiler = filepath.Join("/opt", "android-ndk", "toolchains", "arm-linux-androideabi-4.9", "prebuilt", runtime.GOOS+"-x86_64", "bin", "arm-linux-androideabi-g++")
				}

			case "windows":
				{
					qtPrefix = "C:\\Qt"
					androidPrefix = "C:\\android"
					ndkhost = runtime.GOOS
					compiler = "C:\\android\\android-ndk\\toolchains\\arm-linux-androideabi-4.9\\prebuilt\\windows\\bin\\arm-linux-androideabi-g++.exe"
				}
			}

			var cmd = exec.Command(compiler, "c_main_wrapper.cpp", "-o", filepath.Join(depPath, "libgo.so"), "-I../..", "-L.", "-lgo_base", fmt.Sprintf("--sysroot=%v", filepath.Join(androidPrefix, "android-ndk", "platforms", "android-9", "arch-arm")), "-shared")
			cmd.Dir = depPath
			runCmd(cmd, "predeploy.go_main_wrapper_1")

			runCmd(exec.Command(copyCmd, filepath.Join(depPath, "libgo_base.so"), libPath), "predeploy.cpBase")
			runCmd(exec.Command(copyCmd, filepath.Join(depPath, "libgo.so"), libPath), "predeploy.cpMain")

			var qtLibPath = filepath.Join(qtPrefix, "Qt5.6.0", "5.6", "android_armv7", "lib")
			runCmd(exec.Command(copyCmd, filepath.Join(qtLibPath, "libQt5Widgets.so"), libPath), "predeploy.cpWidgets")
			runCmd(exec.Command(copyCmd, filepath.Join(qtLibPath, "libQt5QuickWidgets.so"), libPath), "predeploy.cpQuickWidgets")
			runCmd(exec.Command(copyCmd, filepath.Join(qtLibPath, "libQt5MultimediaWidgets.so"), libPath), "predeploy.cpMultimediaWidgets")
			runCmd(exec.Command(copyCmd, filepath.Join(qtLibPath, "libQt5Multimedia.so"), libPath), "predeploy.cpMultimedia")
			runCmd(exec.Command(copyCmd, filepath.Join(qtLibPath, "libQt5Network.so"), libPath), "predeploy.cpNetwork")
			runCmd(exec.Command(copyCmd, filepath.Join(qtLibPath, "libQt5AndroidExtras.so"), libPath), "predeploy.cpAndroidExtras")

			var out, err = json.Marshal(&struct {
				Qt                            string `json:"qt"`
				Sdk                           string `json:"sdk"`
				SdkBuildToolsRevision         string `json:"sdkBuildToolsRevision"`
				Ndk                           string `json:"ndk"`
				Toolchainprefix               string `json:"toolchain-prefix"`
				Toolprefix                    string `json:"tool-prefix"`
				Toolchainversion              string `json:"toolchain-version"`
				Ndkhost                       string `json:"ndk-host"`
				Targetarchitecture            string `json:"target-architecture"`
				AndroidExtraLibs              string `json:"android-extra-libs"`
				AndroidPackageSourceDirectory string `json:"android-package-source-directory"`
				Qmlrootpath                   string `json:"qml-root-path"`
				Applicationbinary             string `json:"application-binary"`
			}{
				Qt:  filepath.Join(qtPrefix, "Qt5.6.0", "5.6", "android_armv7"),
				Sdk: filepath.Join(androidPrefix, "android-sdk"),
				SdkBuildToolsRevision: "23.0.3",
				Ndk:                           filepath.Join(androidPrefix, "android-ndk"),
				Toolchainprefix:               "arm-linux-androideabi",
				Toolprefix:                    "arm-linux-androideabi",
				Toolchainversion:              "4.9",
				Ndkhost:                       ndkhost,
				Targetarchitecture:            "armeabi-v7a",
				AndroidExtraLibs:              filepath.Join(depPath, "libgo_base.so"),
				AndroidPackageSourceDirectory: filepath.Join(appPath, "android"),
				Qmlrootpath:                   filepath.Join(appPath, "qml"),
				Applicationbinary:             filepath.Join(depPath, "libgo.so"),
			})
			if err != nil {
				fmt.Println("predeploy.json", string(out), err)
				os.Exit(1)
			}

			utils.Save(filepath.Join(depPath, "android-libgo.so-deployment-settings.json"), string(out))
		}

	case "ios", "ios-simulator":
		{
			utils.MakeFolder(filepath.Join(appPath, buildTarget))

			var buildPath = filepath.Join(depPath, "build")
			utils.MakeFolder(filepath.Join(buildPath, "project.xcodeproj"))
			utils.MakeFolder(filepath.Join(buildPath, "Images.xcassets", "AppIcon.appiconset"))

			//add c_main_wrappers
			utils.Save(filepath.Join(depPath, "c_main_wrapper.cpp"), "#include \"libgo.h\"\nint main(int argc, char *argv[]) { go_main_wrapper(); }")
			//utils.Save(filepath.Join(depPath, "c_main_wrapper_armv7.cpp"), "#include \"libgo_armv7.h\"\nint main(int argc, char *argv[]) { go_main_wrapper(); }")

			utils.Save(filepath.Join(depPath, "gallery_plugin_import.cpp"), iosGalleryPluginImport)
			utils.Save(filepath.Join(depPath, "gallery_qml_plugin_import.cpp"), iosGalleryQmlPluginImport)

			utils.Save(filepath.Join(depPath, "qt.conf"), iosQtConf)

			//build arm64
			var cmd = exec.Command("clang++", "c_main_wrapper.cpp", "gallery_plugin_import.cpp", "gallery_qml_plugin_import.cpp", "-o", "build/main", "-u", "_qt_registerPlatformPlugin", "-Wl,-e,_qt_main_wrapper", "-I../..", "-L.", "-lgo")
			cmd.Args = append(cmd.Args, templater.GetiOSClang(buildTarget, "")...)
			cmd.Dir = depPath
			runCmd(cmd, "predeploy.go_main_wrapper_1")

			//build armv7
			//error						=> ld: Unable to insert branch island. No insertion point available. for architecture armv7
			/*
				cmd = exec.Command("clang++", "c_main_wrapper_armv7.cpp", "gallery_plugin_import.cpp", "gallery_qml_plugin_import.cpp", "-o", "build/main_armv7", "-u", "_qt_registerPlatformPlugin", "-Wl,-e,_qt_main_wrapper", "-I../..", "-L.", "-lgo_armv7")
				cmd.Args = append(cmd.Args, templater.GetiOSClang(buildTarget, "armv7")...)
				cmd.Dir = depPath
				runCmdOptional(cmd, "predeploy.go_main_wrapper_2")
			*/
			//strip arm64 and armv7
			//create fat binary
			//binary size limit	=> https://developer.apple.com/library/ios/documentation/LanguagesUtilities/Conceptual/iTunesConnect_Guide/Chapters/SubmittingTheApp.html

			//create default assets
			utils.Save(filepath.Join(buildPath, "Info.plist"), iosPLIST())
			utils.Save(filepath.Join(buildPath, "Images.xcassets", "AppIcon.appiconset", "Contents.json"), iosAppIcon)
			utils.Save(filepath.Join(buildPath, "LaunchScreen.xib"), iosLaunchScreen())
			utils.Save(filepath.Join(buildPath, "project.xcodeproj", "project.pbxproj"), iosProject())

			runCmd(exec.Command(copyCmd, "/usr/local/Qt5.6.0/5.6/ios/mkspecs/macx-ios-clang/Default-568h@2x.png", buildPath), "predeploy.cpIcon")

			//copy assets from ios folder
			runCmd(exec.Command(copyCmd, "-R", fmt.Sprintf("%v/%v/", appPath, buildTarget), buildPath), "predeploy.cpiOS")
		}

	case "desktop":
		{
			switch runtime.GOOS {
			case "darwin":
				{
					utils.Save(filepath.Join(depPath, fmt.Sprintf("%v.app/Contents/MacOS/%v_sh", appName, appName)), darwinSH())
					utils.Save(filepath.Join(depPath, fmt.Sprintf("%v.app/Contents/MacOS/Info.plist", appName)), darwinPLIST())
					//TODO: icon + plist
				}

			case "linux":
				{
					utils.Save(filepath.Join(depPath, fmt.Sprintf("%v.sh", appName)), linuxSH())
				}

			case "windows":
				{
					//TODO: icon windres
				}
			}
		}
	}
}

func deploy() {

	switch buildTarget {
	case "android":
		{

			var (
				jdkLib   string
				qtPrefix string
				ending   string
			)

			switch runtime.GOOS {
			case "darwin", "linux":
				{

					if runtime.GOOS == "darwin" {
						var version = strings.Split(runCmd(exec.Command("java", "-version"), "deploy.jdk"), "\"")[1]
						jdkLib = fmt.Sprintf("/Library/Java/JavaVirtualMachines/jdk%v.jdk/Contents/Home", version)
					} else {
						jdkLib = "/opt/jdk"
					}

					qtPrefix = "/usr/local"
				}

			case "windows":
				{
					var version = strings.Split(runCmd(exec.Command("java", "-version"), "deploy.jdk"), "\"")[1]
					jdkLib = fmt.Sprintf("C:\\Program Files\\Java\\jdk%v", version)

					qtPrefix = "C:\\Qt"
					ending = ".exe"
				}
			}

			var deploy = exec.Command(filepath.Join(qtPrefix, "Qt5.6.0", "5.6", "android_armv7", "bin", "androiddeployqt"+ending))
			deploy.Args = append(deploy.Args,
				"--input", filepath.Join(depPath, "android-libgo.so-deployment-settings.json"),
				"--output", filepath.Join(depPath, "build"),
				"--deployment", "bundled",
				"--android-platform", "android-22",
				"--jdk", jdkLib,
				"--gradle",
			)

			if ks := utils.Load(filepath.Join(appPath, "android", appName+".keystore")); ks != "" {
				deploy.Args = append(deploy.Args,
					"--sign", filepath.Join(appPath, "android", appName+".keystore"),
					strings.TrimSpace(utils.Load(filepath.Join(appPath, "android", "alias.txt"))),
					"--storepass", strings.TrimSpace(utils.Load(filepath.Join(appPath, "android", "password.txt"))),
				)
			}

			deploy.Dir = filepath.Join(qtPrefix, "Qt5.6.0", "5.6", "android_armv7", "bin")
			deploy.Env = append(deploy.Env, "JAVA_HOME="+jdkLib)

			if runtime.GOOS == "windows" {
				utils.Save(filepath.Join(depPath, "build.bat"), fmt.Sprintf("set JAVA_HOME=%v\r\n%v", jdkLib, strings.Join(deploy.Args, " ")))
				runCmd(exec.Command(filepath.Join(depPath, "build.bat")), "deploy")
				utils.RemoveAll(filepath.Join(depPath, "build.bat"))
			} else {
				runCmd(deploy, "deploy")
			}
		}

	case "ios", "ios-simulator":
		{
			runCmd(exec.Command("xcrun", "xcodebuild", "clean", "build", "CODE_SIGN_IDENTITY=", "CODE_SIGNING_REQUIRED=NO", "CONFIGURATION_BUILD_DIR="+depPath, "-configuration", "Release", "-project", filepath.Join(depPath, "build", "project.xcodeproj")), "deploy")
		}

	case "desktop":
		{
			switch runtime.GOOS {
			case "darwin":
				{
					var deploy = exec.Command("/usr/local/Qt5.6.0/5.6/clang_64/bin/macdeployqt")
					deploy.Args = append(deploy.Args,
						filepath.Join(depPath, fmt.Sprintf("%v.app/", appName)),
						fmt.Sprintf("-qmldir=%v", filepath.Join(appPath, "qml")),
						"-always-overwrite")
					deploy.Dir = "/usr/local/Qt5.6.0/5.6/clang_64/bin/"
					runCmd(deploy, "deploy")
				}

			case "linux":
				{
					var libraryPath string

					for _, dep := range strings.Split(runCmd(exec.Command("ldd", filepath.Join(depPath, appName)), "deploy.ldd"), "\n") {
						if strings.Contains(dep, "libQt5") || strings.Contains(dep, "libicu") {
							var libraryP, libName = filepath.Split(strings.Split(dep, " ")[2])
							libraryPath = libraryP
							runCmd(exec.Command("cp", "-L", filepath.Join(libraryPath, libName), filepath.Join(depPath, libName)), fmt.Sprintf("deploy.%v", libName))
						}
					}

					for _, libName := range []string{"DBus", "XcbQpa", "Quick", "Widgets"} {
						runCmd(exec.Command("cp", "-L", filepath.Join(libraryPath, fmt.Sprintf("libQt5%v.so.5", libName)), filepath.Join(depPath, fmt.Sprintf("libQt5%v.so.5", libName))), fmt.Sprintf("deploy.%v", libName))
					}

					libraryPath = strings.TrimSuffix(libraryPath, "lib/")

					for _, libDir := range []string{"platforms", "platformthemes", "xcbglintegrations"} {
						utils.MakeFolder(filepath.Join(depPath, libDir))
					}

					runCmd(exec.Command("cp", "-R", filepath.Join(libraryPath, "qml/"), depPath), "deploy.qml")
					runCmd(exec.Command("cp", "-L", filepath.Join(libraryPath, "plugins", "platforms", "libqxcb.so"), filepath.Join(depPath, "platforms", "libqxcb.so")), "deploy.qxcb")
					runCmd(exec.Command("cp", "-L", filepath.Join(libraryPath, "plugins", "platformthemes", "libqgtk2.so"), filepath.Join(depPath, "platformthemes", "libqgtk2.so")), "deploy.qgtk2")
					runCmd(exec.Command("cp", "-L", filepath.Join(libraryPath, "plugins", "xcbglintegrations", "libqxcb-glx-integration.so"), filepath.Join(depPath, "xcbglintegrations", "libqxcb-glx-integration.so")), "deploy.qxcb-glx-integration")
				}

			case "windows":
				{
					var deploy = exec.Command("C:\\Qt\\Qt5.6.0\\5.6\\mingw49_32\\bin\\windeployqt.exe")
					deploy.Args = append(deploy.Args,
						filepath.Join(depPath, appName+ending),
						fmt.Sprintf("-qmldir=%v", filepath.Join(appPath, "qml")),
						"-force")
					runCmd(deploy, "deploy")
				}
			}
		}
	}
}

func pastdeploy() {

	switch buildTarget {
	case "android":
		{
			var (
				copyCmd   string
				apkEnding string
			)

			switch runtime.GOOS {
			case "darwin", "linux":
				{
					copyCmd = "cp"
					apkEnding = "apk"
				}

			case "windows":
				{
					copyCmd = "xcopy"
					apkEnding = "apk*"
				}
			}

			if ks := utils.Load(filepath.Join(appPath, "android", appName+".keystore")); ks != "" {
				runCmd(exec.Command(copyCmd, filepath.Join(depPath, "build", "build", "outputs", "apk", "build-release-signed.apk"), filepath.Join(depPath, fmt.Sprintf("%v.%v", appName, apkEnding))), "pastdeploy.release")
			} else {
				runCmd(exec.Command(copyCmd, filepath.Join(depPath, "build", "build", "outputs", "apk", "build-debug.apk"), filepath.Join(depPath, fmt.Sprintf("%v.%v", appName, apkEnding))), "pastdeploy.debug")
			}

			//TODO: copy manifest to android folder and change mindSdkVersion >= 16

		}

	case "ios", "ios-simulator":
		{
			//TODO:
		}

	case "desktop":
		{
			switch runtime.GOOS {
			case "darwin":
				{
					runCmd(exec.Command("mv", filepath.Join(depPath, fmt.Sprintf("%v.app/Contents/MacOS/%v", appName, appName)), filepath.Join(depPath, fmt.Sprintf("%v.app/Contents/MacOS/%v_app", appName, appName))), "pastdeploy.moveApp")
					runCmd(exec.Command("mv", filepath.Join(depPath, fmt.Sprintf("%v.app/Contents/MacOS/%v_sh", appName, appName)), filepath.Join(depPath, fmt.Sprintf("%v.app/Contents/MacOS/%v", appName, appName))), "pastdeploy.moveSh")
				}
			}
		}
	}
}

func cleanup() {
	utils.RemoveAll(filepath.Join(appPath, "qrc.go"))
	utils.RemoveAll(filepath.Join(appPath, "qrc.qrc"))
	utils.RemoveAll(filepath.Join(appPath, "qrc.cpp"))
	utils.RemoveAll(filepath.Join(appPath, "cgo_main_wrapper.go"))
}

func run() {

	switch buildTarget {
	case "android":
		{
			switch runtime.GOOS {
			case "darwin", "linux":
				{
					runCmdOptional(exec.Command("killall", "adb"), "run.killadb")
					//runCmdOptional(exec.Command("/opt/android-sdk/platform-tools/adb", "logcat", "-c"), "run.adblogcat")
					exec.Command("/opt/android-sdk/platform-tools/adb", "install", "-r", filepath.Join(depPath, fmt.Sprintf("%v.apk", appName))).Start()
				}

			case "windows":
				{
					exec.Command("C:\\android\\android-sdk\\platform-tools\\adb.exe", "install", "-r", filepath.Join(depPath, fmt.Sprintf("%v.apk", appName))).Start()
				}
			}
		}

	case /*"ios",*/ "ios-simulator":
		{
			runCmdOptional(exec.Command("xcrun", "instruments", "-w", "iPhone 6s Plus (9.3)#"), "run.boot")
			runCmd(exec.Command("xcrun", "simctl", "uninstall", "booted", filepath.Join(depPath, "main.app")), "run.install")
			runCmd(exec.Command("xcrun", "simctl", "install", "booted", filepath.Join(depPath, "main.app")), "run.install")
			runCmd(exec.Command("xcrun", "simctl", "launch", "booted", fmt.Sprintf("com.identifier.%v", appName)), "run.launch")
		}

	case "desktop":
		{
			switch runtime.GOOS {
			case "darwin":
				{
					runCmdOptional(exec.Command("open", filepath.Join(depPath, fmt.Sprintf("%v.app/", appName))), "run")
				}

			case "linux":
				{
					exec.Command(filepath.Join(depPath, fmt.Sprintf("%v.sh", appName))).Start()
				}

			case "windows":
				{
					exec.Command(filepath.Join(depPath, appName+ending)).Start()
				}
			}
		}
	}
}

func runCmd(cmd *exec.Cmd, name string) string {
	var out, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("\n\n%v\noutput:%s\nerror:%s\n\n", name, out, err)
		cleanup()
		os.Exit(1)
	}
	return string(out)
}

func runCmdOptional(cmd *exec.Cmd, name string) {
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("\n\n%v\noutput:%s\n\n", name, out)
	}
}

//darwin
func darwinSH() string {
	var o = "#!/bin/bash\n"
	o += "cd \"${0%/*}\"\n"
	o += fmt.Sprintf("./%v_app", appName)
	return o
}

func darwinPLIST() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>English</string>

	<key>CFBundleExecutable</key>
	<string>%v</string>

	<key>CFBundleIconFile</key>
	<string>%v.icns</string>

	<key>CFBundleIdentifier</key>
	<string>com.identifier.%v</string>

	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>

	<key>CFBundleName</key>
	<string>%v</string>

	<key>CFBundlePackageType</key>
	<string>APPL</string>

	<key>CFBundleShortVersionString</key>
	<string>1.0.0.0</string>

	<key>CFBundleVersion</key>
	<string>1</string>

	<key>NSHighResolutionCapable</key>
	<string>True</string>
</dict>
</plist>
`, appName, appName, appName, appName)
}

//linux
func linuxSH() string {

	var o = "#!/bin/sh\n"
	o += "appname=`basename $0 | sed s,\\.sh$,,`\n\n"
	o += "dirname=`dirname $0`\n"
	o += "tmp=\"${dirname#?}\"\n\n"
	o += "if [ \"${dirname%$tmp}\" != \"/\" ]; then\n"
	o += "dirname=$PWD/$dirname\n"
	o += "fi\n"

	o += "LD_LIBRARY_PATH=$dirname\n"
	o += "export LD_LIBRARY_PATH\n"

	o += "QML_IMPORT_PATH=$dirname/\"qml\"\n"
	o += "export QML_IMPORT_PATH\n"

	o += "QML2_IMPORT_PATH=$dirname/\"qml\"\n"
	o += "export QML2_IMPORT_PATH\n"

	o += "$dirname/$appname \"$@\"\n"

	return o
}

//ios
func iosPLIST() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>en</string>
	<key>CFBundleExecutable</key>
	<string>main</string>
	<key>CFBundleIdentifier</key>
	<string>com.identifier.%v</string>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>
	<key>CFBundleName</key>
	<string>%v</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleShortVersionString</key>
	<string>1.0</string>
	<key>CFBundleSignature</key>
	<string>????</string>
	<key>CFBundleVersion</key>
	<string>1.0</string>
	<key>LSRequiresIPhoneOS</key>
	<true/>
	<key>UILaunchStoryboardName</key>
	<string>LaunchScreen</string>
	<key>UIRequiredDeviceCapabilities</key>
	<array>
		<string>arm64</string>
	</array>
	<key>UISupportedInterfaceOrientations</key>
	<array>
		<string>UIInterfaceOrientationPortrait</string>
		<string>UIInterfaceOrientationPortraitUpsideDown</string>
		<string>UIInterfaceOrientationLandscapeLeft</string>
		<string>UIInterfaceOrientationLandscapeRight</string>
	</array>
	<key>UISupportedInterfaceOrientations~ipad</key>
	<array>
		<string>UIInterfaceOrientationPortrait</string>
		<string>UIInterfaceOrientationPortraitUpsideDown</string>
		<string>UIInterfaceOrientationLandscapeLeft</string>
		<string>UIInterfaceOrientationLandscapeRight</string>
	</array>
	<key>QtRunLoopIntegrationDisableSeparateStack</key>
	<true/>
</dict>
</plist>
`, appName, appName)
}

func iosLaunchScreen() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
	<document type="com.apple.InterfaceBuilder3.CocoaTouch.XIB" version="3.0" toolsVersion="10117" systemVersion="15E65" targetRuntime="iOS.CocoaTouch" propertyAccessControl="none" useAutolayout="YES" launchScreen="YES" useTraitCollections="YES">
	    <dependencies>
	        <deployment identifier="iOS"/>
	        <plugIn identifier="com.apple.InterfaceBuilder.IBCocoaTouchPlugin" version="10085"/>
	        <capability name="Constraints with non-1.0 multipliers" minToolsVersion="5.1"/>
	    </dependencies>
	    <objects>
	        <placeholder placeholderIdentifier="IBFilesOwner" id="-1" userLabel="File's Owner"/>
	        <placeholder placeholderIdentifier="IBFirstResponder" id="-2" customClass="UIResponder"/>
	        <view contentMode="scaleToFill" id="iN0-l3-epB">
	            <rect key="frame" x="0.0" y="0.0" width="480" height="480"/>
	            <autoresizingMask key="autoresizingMask" widthSizable="YES" heightSizable="YES"/>
	            <subviews>
	                <label opaque="NO" clipsSubviews="YES" userInteractionEnabled="NO" contentMode="left" horizontalHuggingPriority="251" verticalHuggingPriority="251" text="%v" textAlignment="center" lineBreakMode="middleTruncation" baselineAdjustment="alignBaselines" minimumFontSize="18" translatesAutoresizingMaskIntoConstraints="NO" id="kId-c2-rCX">
	                    <rect key="frame" x="20" y="140" width="441" height="43"/>
	                    <fontDescription key="fontDescription" type="boldSystem" pointSize="36"/>
	                    <color key="textColor" red="0.0" green="0.0" blue="0.0" alpha="1" colorSpace="calibratedRGB"/>
	                    <nil key="highlightedColor"/>
	                </label>
	            </subviews>
	            <color key="backgroundColor" white="1" alpha="1" colorSpace="custom" customColorSpace="calibratedWhite"/>
	            <constraints>
	                <constraint firstItem="kId-c2-rCX" firstAttribute="centerY" secondItem="iN0-l3-epB" secondAttribute="bottom" multiplier="1/3" constant="1" id="Kid-kn-2rF"/>
	                <constraint firstAttribute="centerX" secondItem="kId-c2-rCX" secondAttribute="centerX" id="Koa-jz-hwk"/>
	                <constraint firstItem="kId-c2-rCX" firstAttribute="leading" secondItem="iN0-l3-epB" secondAttribute="leading" constant="20" symbolic="YES" id="fvb-Df-36g"/>
	            </constraints>
	            <nil key="simulatedStatusBarMetrics"/>
	            <freeformSimulatedSizeMetrics key="simulatedDestinationMetrics"/>
	            <point key="canvasLocation" x="404" y="445"/>
	        </view>
	    </objects>
	</document>
	`, appName)
}

const iosAppIcon = `{
  "images" : [
    {
      "idiom" : "iphone",
      "size" : "29x29",
      "scale" : "2x"
    },
    {
      "idiom" : "iphone",
      "size" : "29x29",
      "scale" : "3x"
    },
    {
      "idiom" : "iphone",
      "size" : "40x40",
      "scale" : "2x"
    },
    {
      "idiom" : "iphone",
      "size" : "40x40",
      "scale" : "3x"
    },
    {
      "idiom" : "iphone",
      "size" : "60x60",
      "scale" : "2x"
    },
    {
      "idiom" : "iphone",
      "size" : "60x60",
      "scale" : "3x"
    },
    {
      "idiom" : "ipad",
      "size" : "29x29",
      "scale" : "1x"
    },
    {
      "idiom" : "ipad",
      "size" : "29x29",
      "scale" : "2x"
    },
    {
      "idiom" : "ipad",
      "size" : "40x40",
      "scale" : "1x"
    },
    {
      "idiom" : "ipad",
      "size" : "40x40",
      "scale" : "2x"
    },
    {
      "idiom" : "ipad",
      "size" : "76x76",
      "scale" : "1x"
    },
    {
      "idiom" : "ipad",
      "size" : "76x76",
      "scale" : "2x"
    }
  ],
  "info" : {
    "version" : 1,
    "author" : "xcode"
  }
}
`

func iosProject() string {
	return fmt.Sprintf(`// !$*UTF8*$!
{
	archiveVersion = 1;
	classes = {
	};
	objectVersion = 46;
	objects = {

/* Begin PBXBuildFile section */
		254BB84F1B1FD08900C56DE9 /* Images.xcassets in Resources */ = {isa = PBXBuildFile; fileRef = 254BB84E1B1FD08900C56DE9 /* Images.xcassets */; };
		254BB8681B1FD16500C56DE9 /* main in Resources */ = {isa = PBXBuildFile; fileRef = 254BB8671B1FD16500C56DE9 /* main */; };
		25916F411CE65FF600695115 /* LaunchScreen.xib in Resources */ = {isa = PBXBuildFile; fileRef = 25916F401CE65FF600695115 /* LaunchScreen.xib */; };
		25F26AED1CE6675E0045FFBA /* Default-568h@2x.png in Resources */ = {isa = PBXBuildFile; fileRef = 25F26AEC1CE6675E0045FFBA /* Default-568h@2x.png */; };
/* End PBXBuildFile section */

/* Begin PBXFileReference section */
		254BB83E1B1FD08900C56DE9 /* main.app */ = {isa = PBXFileReference; explicitFileType = wrapper.application; includeInIndex = 0; path = main.app; sourceTree = BUILT_PRODUCTS_DIR; };
		254BB8421B1FD08900C56DE9 /* Info.plist */ = {isa = PBXFileReference; lastKnownFileType = text.plist.xml; path = Info.plist; sourceTree = "<group>"; };
		254BB84E1B1FD08900C56DE9 /* Images.xcassets */ = {isa = PBXFileReference; lastKnownFileType = folder.assetcatalog; path = Images.xcassets; sourceTree = "<group>"; };
		254BB8671B1FD16500C56DE9 /* main */ = {isa = PBXFileReference; lastKnownFileType = "compiled.mach-o.executable"; path = main; sourceTree = "<group>"; };
		25916F401CE65FF600695115 /* LaunchScreen.xib */ = {isa = PBXFileReference; fileEncoding = 4; lastKnownFileType = file.xib; path = LaunchScreen.xib; sourceTree = "<group>"; };
		25F26AEC1CE6675E0045FFBA /* Default-568h@2x.png */ = {isa = PBXFileReference; lastKnownFileType = image.png; path = "Default-568h@2x.png"; sourceTree = "<group>"; };
/* End PBXFileReference section */

/* Begin PBXGroup section */
		254BB8351B1FD08900C56DE9 = {
			isa = PBXGroup;
			children = (
				254BB8671B1FD16500C56DE9 /* main */,
				254BB8421B1FD08900C56DE9 /* Info.plist */,
				254BB84E1B1FD08900C56DE9 /* Images.xcassets */,
				25916F401CE65FF600695115 /* LaunchScreen.xib */,
				25F26AEC1CE6675E0045FFBA /* Default-568h@2x.png */,
				254BB83F1B1FD08900C56DE9 /* products */,
			);
			sourceTree = "<group>";
			usesTabs = 0;
		};
		254BB83F1B1FD08900C56DE9 /* products */ = {
			isa = PBXGroup;
			children = (
				254BB83E1B1FD08900C56DE9 /* main.app */,
			);
			name = products;
			sourceTree = "<group>";
		};
/* End PBXGroup section */

/* Begin PBXNativeTarget section */
		254BB83D1B1FD08900C56DE9 /* main */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 254BB8611B1FD08900C56DE9 /* Build configuration list for PBXNativeTarget "main" */;
			buildPhases = (
				254BB83C1B1FD08900C56DE9 /* Resources */,
				259BC5361CE6BA19005B5A05 /* ShellScript */,
			);
			buildRules = (
			);
			dependencies = (
			);
			name = main;
			productName = main;
			productReference = 254BB83E1B1FD08900C56DE9 /* main.app */;
			productType = "com.apple.product-type.application";
		};
/* End PBXNativeTarget section */

/* Begin PBXProject section */
		254BB8361B1FD08900C56DE9 /* Project object */ = {
			isa = PBXProject;
			attributes = {
				LastUpgradeCheck = 0630;
				ORGANIZATIONNAME = Developer;
				TargetAttributes = {
					254BB83D1B1FD08900C56DE9 = {
						CreatedOnToolsVersion = 6.3.1;
					};
				};
			};
			buildConfigurationList = 254BB8391B1FD08900C56DE9 /* Build configuration list for PBXProject "project" */;
			compatibilityVersion = "Xcode 3.2";
			developmentRegion = English;
			hasScannedForEncodings = 0;
			knownRegions = (
				en,
				Base,
			);
			mainGroup = 254BB8351B1FD08900C56DE9;
			productRefGroup = 254BB83F1B1FD08900C56DE9 /* products */;
			projectDirPath = "";
			projectRoot = "";
			targets = (
				254BB83D1B1FD08900C56DE9 /* main */,
			);
		};
/* End PBXProject section */

/* Begin PBXResourcesBuildPhase section */
		254BB83C1B1FD08900C56DE9 /* Resources */ = {
			isa = PBXResourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
				254BB8681B1FD16500C56DE9 /* main in Resources */,
				25F26AED1CE6675E0045FFBA /* Default-568h@2x.png in Resources */,
				25916F411CE65FF600695115 /* LaunchScreen.xib in Resources */,
				254BB84F1B1FD08900C56DE9 /* Images.xcassets in Resources */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
/* End PBXResourcesBuildPhase section */

/* Begin PBXShellScriptBuildPhase section */
		259BC5361CE6BA19005B5A05 /* ShellScript */ = {
			isa = PBXShellScriptBuildPhase;
			buildActionMask = 2147483647;
			files = (
			);
			inputPaths = (
				"$(TARGET_BUILD_DIR)/$(EXECUTABLE_PATH)",
			);
			outputPaths = (
			);
			runOnlyForDeploymentPostprocessing = 0;
			shellPath = /bin/sh;
			shellScript = "cp %v/qt.conf $CODESIGNING_FOLDER_PATH/qt.conf;  test -d $CODESIGNING_FOLDER_PATH/qt_qml && rm -r $CODESIGNING_FOLDER_PATH/qt_qml;  mkdir -p $CODESIGNING_FOLDER_PATH/qt_qml &&  for p in /usr/local/Qt5.6.0/5.6/ios/qml; do rsync -r --exclude='*.a' --exclude='*.prl' --exclude='*.qmltypes'  $p/ $CODESIGNING_FOLDER_PATH/qt_qml; done";
			showEnvVarsInLog = 0;
		};
/* End PBXShellScriptBuildPhase section */

/* Begin XCBuildConfiguration section */
		254BB8601B1FD08900C56DE9 /* Release */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				ALWAYS_SEARCH_USER_PATHS = NO;
				CLANG_CXX_LANGUAGE_STANDARD = "gnu++0x";
				CLANG_CXX_LIBRARY = "libc++";
				CLANG_ENABLE_MODULES = YES;
				CLANG_ENABLE_OBJC_ARC = YES;
				CLANG_WARN_BOOL_CONVERSION = YES;
				CLANG_WARN_CONSTANT_CONVERSION = YES;
				CLANG_WARN_DIRECT_OBJC_ISA_USAGE = YES_ERROR;
				CLANG_WARN_EMPTY_BODY = YES;
				CLANG_WARN_ENUM_CONVERSION = YES;
				CLANG_WARN_INT_CONVERSION = YES;
				CLANG_WARN_OBJC_ROOT_CLASS = YES_ERROR;
				CLANG_WARN_UNREACHABLE_CODE = YES;
				CLANG_WARN__DUPLICATE_METHOD_MATCH = YES;
				"CODE_SIGN_IDENTITY[sdk=iphoneos*]" = "iPhone Developer";
				COPY_PHASE_STRIP = NO;
				DEBUG_INFORMATION_FORMAT = "dwarf-with-dsym";
				ENABLE_NS_ASSERTIONS = NO;
				ENABLE_STRICT_OBJC_MSGSEND = YES;
				GCC_C_LANGUAGE_STANDARD = gnu99;
				GCC_NO_COMMON_BLOCKS = YES;
				GCC_WARN_64_TO_32_BIT_CONVERSION = YES;
				GCC_WARN_ABOUT_RETURN_TYPE = YES_ERROR;
				GCC_WARN_UNDECLARED_SELECTOR = YES;
				GCC_WARN_UNINITIALIZED_AUTOS = YES_AGGRESSIVE;
				GCC_WARN_UNUSED_FUNCTION = YES;
				GCC_WARN_UNUSED_VARIABLE = YES;
				IPHONEOS_DEPLOYMENT_TARGET = 9.3;
				MTL_ENABLE_DEBUG_INFO = NO;
				SDKROOT = iphoneos;
				TARGETED_DEVICE_FAMILY = "1,2";
				VALIDATE_PRODUCT = YES;
			};
			name = Release;
		};
		254BB8631B1FD08900C56DE9 /* Release */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				ASSETCATALOG_COMPILER_APPICON_NAME = AppIcon;
				INFOPLIST_FILE = Info.plist;
				LD_RUNPATH_SEARCH_PATHS = "$(inherited) @executable_path/Frameworks";
				PRODUCT_NAME = "$(TARGET_NAME)";
			};
			name = Release;
		};
/* End XCBuildConfiguration section */

/* Begin XCConfigurationList section */
		254BB8391B1FD08900C56DE9 /* Build configuration list for PBXProject "project" */ = {
			isa = XCConfigurationList;
			buildConfigurations = (
				254BB8601B1FD08900C56DE9 /* Release */,
			);
			defaultConfigurationIsVisible = 0;
			defaultConfigurationName = Release;
		};
		254BB8611B1FD08900C56DE9 /* Build configuration list for PBXNativeTarget "main" */ = {
			isa = XCConfigurationList;
			buildConfigurations = (
				254BB8631B1FD08900C56DE9 /* Release */,
			);
			defaultConfigurationIsVisible = 0;
			defaultConfigurationName = Release;
		};
/* End XCConfigurationList section */
	};
	rootObject = 254BB8361B1FD08900C56DE9 /* Project object */;
}
`, depPath)
}

const (
	iosGalleryPluginImport = `#include <QtPlugin>
Q_IMPORT_PLUGIN(AVFServicePlugin)
Q_IMPORT_PLUGIN(AVFMediaPlayerServicePlugin)
Q_IMPORT_PLUGIN(AudioCaptureServicePlugin)
Q_IMPORT_PLUGIN(CoreAudioPlugin)
Q_IMPORT_PLUGIN(QM3uPlaylistPlugin)
Q_IMPORT_PLUGIN(QDDSPlugin)
Q_IMPORT_PLUGIN(QICNSPlugin)
Q_IMPORT_PLUGIN(QICOPlugin)
Q_IMPORT_PLUGIN(QTgaPlugin)
Q_IMPORT_PLUGIN(QTiffPlugin)
Q_IMPORT_PLUGIN(QWbmpPlugin)
Q_IMPORT_PLUGIN(QWebpPlugin)
Q_IMPORT_PLUGIN(QQmlDebuggerServiceFactory)
Q_IMPORT_PLUGIN(QQmlInspectorServiceFactory)
Q_IMPORT_PLUGIN(QLocalClientConnectionFactory)
Q_IMPORT_PLUGIN(QQmlNativeDebugConnectorFactory)
Q_IMPORT_PLUGIN(QQmlProfilerServiceFactory)
Q_IMPORT_PLUGIN(QQmlDebugServerFactory)
Q_IMPORT_PLUGIN(QTcpServerConnectionFactory)
`

	iosGalleryQmlPluginImport = `#include <QtPlugin>
Q_IMPORT_PLUGIN(QtQuick2Plugin)
Q_IMPORT_PLUGIN(QMultimediaDeclarativeModule)
Q_IMPORT_PLUGIN(QtQuickLayoutsPlugin)
Q_IMPORT_PLUGIN(QtQuick2DialogsPlugin)
Q_IMPORT_PLUGIN(QtQuickControlsPlugin)
Q_IMPORT_PLUGIN(QmlFolderListModelPlugin)
Q_IMPORT_PLUGIN(QmlSettingsPlugin)
Q_IMPORT_PLUGIN(QtQuick2DialogsPrivatePlugin)
Q_IMPORT_PLUGIN(QtQuick2WindowPlugin)
Q_IMPORT_PLUGIN(QtQmlModelsPlugin)
Q_IMPORT_PLUGIN(QtQuickExtrasPlugin)
Q_IMPORT_PLUGIN(QtGraphicalEffectsPlugin)
`

	iosQtConf = `[Paths]
Imports = qt_qml
Qml2Imports = qt_qml
`
)
