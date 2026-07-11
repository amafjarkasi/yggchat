#!/usr/bin/env python3
"""Generate a zoomed-in logo for Yggdrasil Mesh Chat with text above and below"""

from PIL import Image, ImageDraw, ImageFont
import math

def create_logo(size=800):
    """Create a zoomed-in logo with icon centered, text above and below"""
    
    # Create image with dark background
    img = Image.new('RGB', (size, size), (26, 27, 38))
    draw = ImageDraw.Draw(img)
    
    # Colors - Tokyo Night palette
    bg = (26, 27, 38)
    dark = (17, 17, 27)
    blue = (122, 162, 247)
    purple = (187, 154, 247)
    green = (158, 206, 106)
    amber = (224, 175, 104)
    text_light = (169, 177, 214)
    white = (248, 248, 242)
    
    cx, cy = size // 2, size // 2
    
    # Load fonts
    try:
        font_paths = [
            "C:/Windows/Fonts/arialbd.ttf",
            "C:/Windows/Fonts/arial.ttf",
            "C:/Windows/Fonts/segoeui.ttf",
            "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
        ]
        
        font_title = None
        font_sub = None
        
        for path in font_paths:
            try:
                font_title = ImageFont.truetype(path, 72)
                font_sub = ImageFont.truetype(path, 48)
                break
            except:
                continue
        
        if font_title is None:
            font_title = ImageFont.load_default()
            font_sub = ImageFont.load_default()
            
    except:
        font_title = ImageFont.load_default()
        font_sub = ImageFont.load_default()
    
    # === Draw "YGGDRASIL" text at top ===
    title = "YGGDRASIL"
    title_bbox = draw.textbbox((0, 0), title, font=font_title)
    title_w = title_bbox[2] - title_bbox[0]
    title_h = title_bbox[3] - title_bbox[1]
    title_x = cx - title_w // 2
    title_y = 100
    
    # Draw title with glow effect
    for offset in range(3, 0, -1):
        glow_color = (blue[0]//4, blue[1]//4, blue[2]//4)
        draw.text((title_x - offset, title_y), title, fill=glow_color, font=font_title)
        draw.text((title_x + offset, title_y), title, fill=glow_color, font=font_title)
    
    draw.text((title_x, title_y), title, fill=blue, font=font_title)
    
    # === Draw the tree icon (zoomed in, centered) ===
    
    # Tree trunk - centered
    trunk_x = cx
    trunk_top = cy - 100
    trunk_bottom = cy + 80
    trunk_width = 12
    
    # Draw trunk with gradient effect
    for y in range(trunk_top, trunk_bottom):
        progress = (y - trunk_top) / (trunk_bottom - trunk_top)
        width = int(8 + progress * 10)
        color = (
            int(blue[0] * (1 - progress * 0.3)),
            int(blue[1] * (1 - progress * 0.3)),
            int(blue[2] * (1 - progress * 0.3))
        )
        draw.line([(trunk_x - width//2, y), (trunk_x + width//2, y)], fill=color)
    
    # === Branches (larger, more visible) ===
    branch_data = [
        # (angle, length, y_offset, thickness)
        (-65, 130, -50, 7),
        (-40, 150, -70, 7),
        (-15, 110, -90, 6),
        (0, 100, -100, 6),      # top
        (15, 110, -90, 6),
        (40, 150, -70, 7),
        (65, 130, -50, 7),
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
        
        # Draw larger node at end
        node_r = 10
        draw.ellipse(
            [end_x - node_r, end_y - node_r, end_x + node_r, end_y + node_r],
            fill=green,
            outline=white,
            width=3
        )
    
    # === Roots (larger) ===
    root_data = [
        (-35, 100, 6),
        (-15, 90, 5),
        (15, 90, 5),
        (35, 100, 6),
    ]
    
    for angle, length, thickness in root_data:
        start_x = trunk_x
        start_y = trunk_bottom - 10
        
        angle_rad = math.radians(angle + 90)
        end_x = start_x + math.cos(angle_rad) * length
        end_y = start_y + math.sin(angle_rad) * length * 0.6
        
        draw.line([(start_x, start_y), (end_x, end_y)], fill=blue, width=thickness)
    
    # === Network mesh lines ===
    mesh_pairs = [(0, 2), (1, 3), (2, 4), (3, 5), (4, 6), (0, 6), (1, 5)]
    
    for i, j in mesh_pairs:
        if i < len(branch_points) and j < len(branch_points):
            x1, y1 = branch_points[i]
            x2, y2 = branch_points[j]
            
            # Draw dashed line
            dash_length = 10
            gap_length = 8
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
                    
                    draw.line([(sx, sy), (ex, ey)], fill=blue, width=2)
                    pos += dash_length + gap_length
    
    # === Lightning bolt (larger) ===
    bolt_cx = cx
    bolt_cy = cy + 40
    bolt_size = 40
    
    bolt_points = [
        (bolt_cx - 12, bolt_cy - bolt_size),
        (bolt_cx + 18, bolt_cy - bolt_size),
        (bolt_cx + 5, bolt_cy - 8),
        (bolt_cx + 22, bolt_cy - 8),
        (bolt_cx - 10, bolt_cy + bolt_size),
        (bolt_cx + 8, bolt_cy + 10),
        (bolt_cx - 15, bolt_cy + 10),
    ]
    draw.polygon(bolt_points, fill=amber)
    
    # === Draw "MESH CHAT" text at bottom ===
    subtitle = "MESH CHAT"
    subtitle_bbox = draw.textbbox((0, 0), subtitle, font=font_sub)
    subtitle_w = subtitle_bbox[2] - subtitle_bbox[0]
    subtitle_x = cx - subtitle_w // 2
    subtitle_y = size - 150
    
    # Draw subtitle with glow effect
    for offset in range(2, 0, -1):
        glow_color = (purple[0]//4, purple[1]//4, purple[2]//4)
        draw.text((subtitle_x - offset, subtitle_y), subtitle, fill=glow_color, font=font_sub)
        draw.text((subtitle_x + offset, subtitle_y), subtitle, fill=glow_color, font=font_sub)
    
    draw.text((subtitle_x, subtitle_y), subtitle, fill=purple, font=font_sub)
    
    # === Decorative horizontal lines ===
    line_y_top = title_y + title_h + 30
    line_y_bottom = subtitle_y - 30
    line_width = 200
    
    # Top line
    draw.line([(cx - line_width, line_y_top), (cx + line_width, line_y_top)], 
              fill=(*blue[:3], 100), width=2)
    
    # Bottom line
    draw.line([(cx - line_width, line_y_bottom), (cx + line_width, line_y_bottom)], 
              fill=(*purple[:3], 100), width=2)
    
    # === Small decorative dots ===
    dot_positions = [
        (cx - line_width - 10, line_y_top),
        (cx + line_width + 10, line_y_top),
        (cx - line_width - 10, line_y_bottom),
        (cx + line_width + 10, line_y_bottom),
    ]
    
    for dx, dy in dot_positions:
        draw.ellipse([dx-4, dy-4, dx+4, dy+4], fill=green)
    
    return img


def create_rounded_logo(size=800):
    """Create a circular version of the logo"""
    
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
    print("Generating Yggdrasil Mesh Chat logo (zoomed version)...")
    
    # Generate square logo
    square_logo = create_logo(800)
    square_logo.save("logo_square.png", "PNG")
    print("Saved: logo_square.png (800x800)")
    
    # Generate circular logo
    round_logo = create_rounded_logo(800)
    round_logo.save("logo.png", "PNG")
    print("Saved: logo.png (800x800 circular)")
    
    # Generate small circular logo
    small_round = create_rounded_logo(160)
    small_round.save("logo_small.png", "PNG")
    print("Saved: logo_small.png (160x160)")
    
    # Generate JPG version
    jpg_bg = Image.new('RGB', (800, 800), (26, 27, 38))
    jpg_bg.paste(round_logo, mask=round_logo.split()[3])
    jpg_bg.save("logo.jpg", "JPEG", quality=95)
    print("Saved: logo.jpg")
    
    print("\nLogo generation complete!")


if __name__ == "__main__":
    main()
