#!/usr/bin/env python3
"""Generate a clean, modern logo for Yggdrasil Mesh Chat"""

from PIL import Image, ImageDraw, ImageFont
import math

def create_logo(size=800):
    """Create a clean, modern Yggdrasil Mesh Chat logo"""
    
    # Create image with dark background
    img = Image.new('RGB', (size, size), (26, 27, 38))
    draw = ImageDraw.Draw(img)
    
    # Colors - Tokyo Night palette
    bg = (26, 27, 38)
    dark = (22, 23, 33)
    blue = (122, 162, 247)
    purple = (187, 154, 247)
    green = (158, 206, 106)
    amber = (224, 175, 104)
    text_light = (169, 177, 214)
    white = (248, 248, 242)
    
    cx, cy = size // 2, size // 2
    
    # Draw rounded rectangle background
    margin = 40
    corner_radius = 60
    draw.rounded_rectangle(
        [margin, margin, size - margin, size - margin],
        radius=corner_radius,
        fill=dark
    )
    
    # Draw subtle border
    draw.rounded_rectangle(
        [margin, margin, size - margin, size - margin],
        radius=corner_radius,
        outline=blue,
        width=3
    )
    
    # === Draw stylized tree ===
    
    # Trunk - clean vertical line
    trunk_x = cx
    trunk_top = cy - 140
    trunk_bottom = cy + 100
    trunk_width = 10
    
    # Draw trunk with gradient effect (thicker at bottom)
    for y in range(trunk_top, trunk_bottom):
        progress = (y - trunk_top) / (trunk_bottom - trunk_top)
        width = int(6 + progress * 8)
        color = (
            int(blue[0] * (1 - progress * 0.3)),
            int(blue[1] * (1 - progress * 0.3)),
            int(blue[2] * (1 - progress * 0.3))
        )
        draw.line([(trunk_x - width//2, y), (trunk_x + width//2, y)], fill=color)
    
    # === Branches ===
    # Clean, symmetrical branch structure
    
    branch_data = [
        # (angle, length, y_offset, thickness)
        (-70, 100, -80, 5),
        (-45, 120, -100, 5),
        (-20, 90, -120, 4),
        (0, 80, -140, 4),      # top
        (20, 90, -120, 4),
        (45, 120, -100, 5),
        (70, 100, -80, 5),
    ]
    
    branch_points = []
    
    for angle, length, y_offset, thickness in branch_data:
        start_x = trunk_x
        start_y = cy + y_offset
        
        angle_rad = math.radians(angle - 90)
        end_x = start_x + math.cos(angle_rad) * length
        end_y = start_y + math.sin(angle_rad) * length
        
        # Draw branch
        draw.line([(start_x, start_y), (end_x, end_y)], fill=purple, width=thickness)
        
        # Store endpoint for nodes
        branch_points.append((end_x, end_y))
        
        # Draw small leaf/node at end
        node_r = 6
        draw.ellipse(
            [end_x - node_r, end_y - node_r, end_x + node_r, end_y + node_r],
            fill=green,
            outline=white,
            width=2
        )
    
    # === Roots ===
    root_data = [
        (-30, 80, 5),
        (-10, 70, 4),
        (10, 70, 4),
        (30, 80, 5),
    ]
    
    for angle, length, thickness in root_data:
        start_x = trunk_x
        start_y = trunk_bottom - 20
        
        angle_rad = math.radians(angle + 90)
        end_x = start_x + math.cos(angle_rad) * length
        end_y = start_y + math.sin(angle_rad) * length * 0.6
        
        draw.line([(start_x, start_y), (end_x, end_y)], fill=blue, width=thickness)
    
    # === Network mesh lines ===
    # Connect some nodes with subtle lines
    
    mesh_pairs = [(0, 2), (1, 3), (2, 4), (3, 5), (4, 6), (0, 6), (1, 5)]
    
    for i, j in mesh_pairs:
        if i < len(branch_points) and j < len(branch_points):
            x1, y1 = branch_points[i]
            x2, y2 = branch_points[j]
            
            # Draw dashed line
            dash_length = 8
            gap_length = 6
            dx = x2 - x1
            dy = y2 - y1
            dist = math.sqrt(dx*dx + dy*dy)
            
            if dist > 0:
                dx /= dist
                dy /= dist
                
                pos = 0
                while pos < dist:
                    sx = x1 + dx * pos
                    sy = y1 + dy * pos
                    end_pos = min(pos + dash_length, dist)
                    ex = x1 + dx * end_pos
                    ey = y1 + dy * end_pos
                    
                    # Fade based on position
                    alpha = int(80 * (1 - pos/dist * 0.5))
                    color = (blue[0], blue[1], blue[2])
                    
                    draw.line([(sx, sy), (ex, ey)], fill=color, width=2)
                    pos += dash_length + gap_length
    
    # === Lightning bolt symbol ===
    bolt_cx = cx
    bolt_cy = cy + 60
    bolt_size = 30
    
    bolt_points = [
        (bolt_cx - 10, bolt_cy - bolt_size),
        (bolt_cx + 15, bolt_cy - bolt_size),
        (bolt_cx + 3, bolt_cy - 5),
        (bolt_cx + 18, bolt_cy - 5),
        (bolt_cx - 8, bolt_cy + bolt_size),
        (bolt_cx + 7, bolt_cy + 8),
        (bolt_cx - 12, bolt_cy + 8),
    ]
    draw.polygon(bolt_points, fill=amber)
    
    # === Text ===
    
    # Load fonts
    try:
        # Try different font paths
        font_paths = [
            "C:/Windows/Fonts/arial.ttf",
            "C:/Windows/Fonts/segoeui.ttf",
            "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
            "arial.ttf",
        ]
        
        font_large = None
        font_small = None
        
        for path in font_paths:
            try:
                font_large = ImageFont.truetype(path, 48)
                font_small = ImageFont.truetype(path, 32)
                break
            except:
                continue
        
        if font_large is None:
            font_large = ImageFont.load_default()
            font_small = ImageFont.load_default()
            
    except:
        font_large = ImageFont.load_default()
        font_small = ImageFont.load_default()
    
    # Title text "YGGDRASIL" - centered at top
    title = "YGGDRASIL"
    title_bbox = draw.textbbox((0, 0), title, font=font_large)
    title_w = title_bbox[2] - title_bbox[0]
    title_x = cx - title_w // 2
    title_y = margin + 40
    
    draw.text((title_x, title_y), title, fill=blue, font=font_large)
    
    # Subtitle "MESH CHAT" - centered at bottom
    subtitle = "MESH CHAT"
    subtitle_bbox = draw.textbbox((0, 0), subtitle, font=font_small)
    subtitle_w = subtitle_bbox[2] - subtitle_bbox[0]
    subtitle_x = cx - subtitle_w // 2
    subtitle_y = size - margin - 80
    
    draw.text((subtitle_x, subtitle_y), subtitle, fill=purple, font=font_small)
    
    # === Decorative elements ===
    
    # Small dots in corners
    dot_positions = [
        (margin + 30, margin + 30),
        (size - margin - 30, margin + 30),
        (margin + 30, size - margin - 30),
        (size - margin - 30, size - margin - 30),
    ]
    
    for dx, dy in dot_positions:
        draw.ellipse([dx-4, dy-4, dx+4, dy+4], fill=green)
    
    return img


def create_rounded_logo(size=800):
    """Create a circular version of the logo"""
    
    # First create the square logo
    square_logo = create_logo(size)
    
    # Create circular mask
    mask = Image.new('L', (size, size), 0)
    mask_draw = ImageDraw.Draw(mask)
    mask_draw.ellipse([0, 0, size-1, size-1], fill=255)
    
    # Apply mask
    output = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    output.paste(square_logo, mask=mask)
    
    return output


def main():
    """Generate and save the logos"""
    print("Generating Yggdrasil Mesh Chat logo...")
    
    # Generate square logo (for GitHub social preview)
    square_logo = create_logo(800)
    square_logo.save("logo_square.png", "PNG")
    print("Saved: logo_square.png (800x800)")
    
    # Generate circular logo (for README)
    round_logo = create_rounded_logo(800)
    round_logo.save("logo.png", "PNG")
    print("Saved: logo.png (800x800 circular)")
    
    # Generate small circular logo
    small_round = create_rounded_logo(160)
    small_round.save("logo_small.png", "PNG")
    print("Saved: logo_small.png (160x160)")
    
    # Generate JPG version (with dark background)
    jpg_bg = Image.new('RGB', (800, 800), (26, 27, 38))
    jpg_bg.paste(round_logo, mask=round_logo.split()[3])
    jpg_bg.save("logo.jpg", "JPEG", quality=95)
    print("Saved: logo.jpg")
    
    print("\nLogo generation complete!")
    print("\nFiles created:")
    print("  - logo.png      : Circular logo for README (transparent bg)")
    print("  - logo_square.png : Square version for social previews")
    print("  - logo_small.png  : Small 160x160 version")
    print("  - logo.jpg       : JPG version (dark bg)")


if __name__ == "__main__":
    main()
