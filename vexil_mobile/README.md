How To Build :
cd vexil_app
flutter build apk --release

Attention:
if you want to rewrite vexil_go, should use the codes to update gomobile before rebuild flutter

cd vexil_go\bridge
gomobile bind -target=android -androidapi 21 -o ../../vexil_app/android/app/libs/vexil.aar .
cd vexil_mobile\vexil_app
flutter build apk --release
