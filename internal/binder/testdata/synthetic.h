#ifndef SYNTH_H
#define SYNTH_H

enum Color {
	RED,
	GREEN = 10,
	BLUE
};

struct Point {
	int x;
	int y;
};

typedef struct Point Point_t;

void draw_point(Point_t p, int color);
const char *get_name(void);

#endif
