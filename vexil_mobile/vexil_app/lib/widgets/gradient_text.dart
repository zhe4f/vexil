import 'package:flutter/material.dart';

class GradientText extends StatelessWidget {
  final String text;
  final double fontSize;
  final FontWeight fontWeight;
  final List<Color> colors;

  const GradientText(
    this.text, {
    super.key,
    this.fontSize = 20,
    this.fontWeight = FontWeight.w300,
    this.colors = const [Color(0xFF5B7CFA), Color(0xFF7C8CF8)],
  });

  @override
  Widget build(BuildContext context) {
    return ShaderMask(
      shaderCallback: (bounds) => LinearGradient(
        colors: colors,
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
      ).createShader(bounds),
      child: Text(
        text,
        style: TextStyle(
          fontSize: fontSize,
          fontWeight: fontWeight,
          color: Colors.white,
          letterSpacing: 4,
        ),
      ),
    );
  }
}