"""Drumstick outline generator for the logo drafts.

A stick is described along its axis s (0 = top of the tip bead, L = end of
the butt) with a half-width profile:
  - bead: an acorn-shaped ellipse at the tip
  - neck: the thin point just below the bead
  - taper: a long, gentle curve from the neck to the full shaft width
  - shaft: constant width, ending in a rounded butt
Points are mapped into the artboard with the stick's tip position and angle.
"""
import math
import sys


def stick_paths(tip, angle_deg, length, half_width=7.5, bead_rx=4.2, bead_ry=7.2, neck_width=1.9):
    """Return (body_d, bead_d) SVG path data for one stick.

    tip: (x, y) of the top of the bead; angle_deg: tilt from vertical,
    positive leans the stick so the butt goes to the right.
    """
    a = math.radians(angle_deg)
    d = (math.sin(a), math.cos(a))          # along the stick, tip -> butt
    n = (-d[1], d[0])                       # across the stick

    def P(s, w):
        return (tip[0] + s * d[0] + w * n[0], tip[1] + s * d[1] + w * n[1])

    def f(p):
        return f"{p[0]:.2f} {p[1]:.2f}"

    L, hw = length, half_width
    neck_s, neck_w = bead_ry * 2 - 1.0, neck_width  # overlaps the bead slightly
    full_s = 0.52 * L                       # where the taper reaches full width
    body = (
        f"M{f(P(neck_s, neck_w))} "
        f"C{f(P(neck_s + 0.2 * L, neck_w + 0.6))} {f(P(full_s - 0.13 * L, hw))} {f(P(full_s, hw))} "
        f"L{f(P(L - hw, hw))} "
        f"A{hw} {hw} 0 0 0 {f(P(L - hw, -hw))} "
        f"L{f(P(full_s, -hw))} "
        f"C{f(P(full_s - 0.13 * L, -hw))} {f(P(neck_s + 0.2 * L, -neck_w - 0.6))} {f(P(neck_s, -neck_w))} Z"
    )
    rot = angle_deg  # ellipse x-axis rotation: rx runs across the stick
    top, bottom = P(0, 0), P(2 * bead_ry, 0)
    bead = f"M{f(top)} A{bead_rx} {bead_ry} {-rot:.2f} 1 1 {f(bottom)} A{bead_rx} {bead_ry} {-rot:.2f} 1 1 {f(top)} Z"
    return body, bead


def crossed(center_x, top_y, length, angle=30.0, cross_at=0.47):
    """Two sticks forming an X, tips at the top, symmetric around center_x."""
    a = math.radians(angle)
    c = cross_at * length
    py = top_y + c * math.cos(a)
    left_tip = (center_x - c * math.sin(a), top_y)    # leans to bottom-right
    right_tip = (center_x + c * math.sin(a), top_y)   # leans to bottom-left
    return stick_paths(left_tip, angle, length), stick_paths(right_tip, -angle, length), py


if __name__ == "__main__":
    cx, top, length = (float(v) for v in sys.argv[1:4])
    angle = float(sys.argv[4]) if len(sys.argv) > 4 else 30.0
    (lb, lbead), (rb, rbead), py = crossed(cx, top, length, angle)
    print("LEFT_BODY", lb)
    print("LEFT_BEAD", lbead)
    print("RIGHT_BODY", rb)
    print("RIGHT_BEAD", rbead)
    print("CROSS_Y", round(py, 2))
