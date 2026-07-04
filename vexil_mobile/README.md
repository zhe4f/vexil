How To Build :
<br />
cd vexil_app
<br />
flutter build apk --release
<br />
<br />
Attention:
<br />
if you want to rewrite vexil_go, should use the codes to update gomobile before rebuild flutter
<br />
cd vexil_go\bridge
<br />
gomobile bind -target=android -androidapi 21 -o ../../vexil_app/android/app/libs/vexil.aar .
<br />
cd vexil_mobile\vexil_app
<br />
flutter build apk --release
<br />
