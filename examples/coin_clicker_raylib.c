// Generated Cortex Program
#define CORTEX_FEATURE_ASYNC 1
#define CORTEX_FEATURE_ACTORS 1
#define CORTEX_FEATURE_BLOCKCHAIN 1
#define CORTEX_FEATURE_QOL 1

#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <stdbool.h>
#include <time.h>
#include <string.h>
#include "runtime/core.h"
#include "runtime/game.h"

#include <raylib.h>



double coins = 0.0;
double coins_per_click = 1.0;
double coins_per_second = 0.0;
int total_clicks = 0;
int cursors = 0;
int grandmas = 0;
int farms = 0;
int mines = 0;
int factories = 0;
double cursor_cost = 15.0;
double grandma_cost = 100.0;
double farm_cost = 1100.0;
double mine_cost = 12000.0;
double factory_cost = 130000.0;
void format_coins(double amount, char* buffer) {
    if ((amount >= 1000000000.0)) {
        sprintf(buffer, "%.2f B", (amount / (double)1000000000.0));
} else {
        if ((amount >= 1000000.0)) {
            sprintf(buffer, "%.2f M", (amount / (double)1000000.0));
} else {
            if ((amount >= 1000.0)) {
                sprintf(buffer, "%.2f K", (amount / (double)1000.0));
} else {
                sprintf(buffer, "%.1f", amount);
}

}

}

}

void main() {
    InitWindow(500, 600, "Coin Clicker!");
SetTargetFPS(60);
char buf[128];
char cost_buf[64];
int mx = 0;
int my = 0;
int click = 0;
while ((!WindowShouldClose())) {
        mx = GetMouseX();
my = GetMouseY();
click = IsMouseButtonPressed(MOUSE_LEFT_BUTTON);
coins = (coins + (double)(coins_per_second / (double)60.0));
if ((click == 1)) {
            if ((mx >= 150)) {
                if ((mx <= 350)) {
                    if ((my >= 80)) {
                        if ((my <= 160)) {
                            coins = (coins + coins_per_click);
total_clicks = (total_clicks + (int)1);
}

}

}

}

}

if ((click == 1)) {
            if ((mx >= 20)) {
                if ((mx <= 240)) {
                    if ((my >= 240)) {
                        if ((my <= 280)) {
                            if ((coins >= cursor_cost)) {
                                coins = (coins - cursor_cost);
cursors = (cursors + (int)1);
coins_per_second = (coins_per_second + (double)0.1);
cursor_cost = (cursor_cost * (double)1.15);
}

}

}

}

}

}

if ((click == 1)) {
            if ((mx >= 260)) {
                if ((mx <= 480)) {
                    if ((my >= 240)) {
                        if ((my <= 280)) {
                            if ((coins >= grandma_cost)) {
                                coins = (coins - grandma_cost);
grandmas = (grandmas + (int)1);
coins_per_second = (coins_per_second + (double)1.0);
grandma_cost = (grandma_cost * (double)1.15);
}

}

}

}

}

}

if ((click == 1)) {
            if ((mx >= 20)) {
                if ((mx <= 240)) {
                    if ((my >= 300)) {
                        if ((my <= 340)) {
                            if ((coins >= farm_cost)) {
                                coins = (coins - farm_cost);
farms = (farms + (int)1);
coins_per_second = (coins_per_second + (double)8.0);
farm_cost = (farm_cost * (double)1.15);
}

}

}

}

}

}

if ((click == 1)) {
            if ((mx >= 260)) {
                if ((mx <= 480)) {
                    if ((my >= 300)) {
                        if ((my <= 340)) {
                            if ((coins >= mine_cost)) {
                                coins = (coins - mine_cost);
mines = (mines + (int)1);
coins_per_second = (coins_per_second + (double)47.0);
mine_cost = (mine_cost * (double)1.15);
}

}

}

}

}

}

if ((click == 1)) {
            if ((mx >= 140)) {
                if ((mx <= 360)) {
                    if ((my >= 360)) {
                        if ((my <= 400)) {
                            if ((coins >= factory_cost)) {
                                coins = (coins - factory_cost);
factories = (factories + (int)1);
coins_per_second = (coins_per_second + (double)260.0);
factory_cost = (factory_cost * (double)1.15);
}

}

}

}

}

}

if ((click == 1)) {
            if ((mx >= 150)) {
                if ((mx <= 350)) {
                    if ((my >= 520)) {
                        if ((my <= 560)) {
                            coins = 0;
coins_per_click = 1.0;
coins_per_second = 0.0;
total_clicks = 0;
cursors = 0;
grandmas = 0;
farms = 0;
mines = 0;
factories = 0;
cursor_cost = 15.0;
grandma_cost = 100.0;
farm_cost = 1100.0;
mine_cost = 12000.0;
factory_cost = 130000.0;
}

}

}

}

}

BeginDrawing();
ClearBackground(DARKBLUE);
DrawText("COIN CLICKER", 130, 20, 40, GOLD);
format_coins(coins, buf);
DrawText(buf, 180, 170, 30, YELLOW);
sprintf(buf, "%.1f/sec | Clicks: %d", coins_per_second, total_clicks);
DrawText(buf, 120, 200, 18, LIGHTGRAY);
DrawRectangle(150, 80, 200, 80, GOLD);
DrawRectangleLines(150, 80, 200, 80, DARKGRAY);
DrawText("CLICK!", 210, 105, 30, WHITE);
DrawText("UPGRADES", 190, 420, 25, WHITE);
format_coins(cursor_cost, cost_buf);
sprintf(buf, "Cursor (%s)", cost_buf);
DrawRectangle(20, 240, 220, 40, BROWN);
DrawRectangleLines(20, 240, 220, 40, DARKGRAY);
DrawText(buf, 30, 250, 18, WHITE);
format_coins(grandma_cost, cost_buf);
sprintf(buf, "Grandma (%s)", cost_buf);
DrawRectangle(260, 240, 220, 40, ORANGE);
DrawRectangleLines(260, 240, 220, 40, DARKGRAY);
DrawText(buf, 270, 250, 18, WHITE);
format_coins(farm_cost, cost_buf);
sprintf(buf, "Farm (%s)", cost_buf);
DrawRectangle(20, 300, 220, 40, GREEN);
DrawRectangleLines(20, 300, 220, 40, DARKGRAY);
DrawText(buf, 30, 310, 18, WHITE);
format_coins(mine_cost, cost_buf);
sprintf(buf, "Mine (%s)", cost_buf);
DrawRectangle(260, 300, 220, 40, GRAY);
DrawRectangleLines(260, 300, 220, 40, DARKGRAY);
DrawText(buf, 270, 310, 18, WHITE);
format_coins(factory_cost, cost_buf);
sprintf(buf, "Factory (%s)", cost_buf);
DrawRectangle(140, 360, 220, 40, PURPLE);
DrawRectangleLines(140, 360, 220, 40, DARKGRAY);
DrawText(buf, 150, 370, 18, WHITE);
sprintf(buf, "Owned: %d | %d | %d | %d | %d", cursors, grandmas, farms, mines, factories);
DrawText(buf, 80, 450, 18, LIGHTGRAY);
DrawText("Click the gold button to earn coins!", 80, 480, 16, GRAY);
DrawRectangle(150, 520, 200, 40, MAROON);
DrawRectangleLines(150, 520, 200, 40, DARKGRAY);
DrawText("Reset", 215, 530, 18, WHITE);
EndDrawing();
}
CloseWindow();
}

