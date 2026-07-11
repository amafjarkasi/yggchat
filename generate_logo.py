#!/usr/bin/env python3
"""Generate a modern logo for Yggdrasil Mesh Chat"""

from PIL import Image, ImageDraw, ImageFont
import math

def create_logo(size=800):
    """Create a modern Yggdrasil Mesh Chat logo"""
    
    # Create image with transparent background
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    
    # Tokyo Night inspired colors
    bg_color = (26, 27, 38, 255)        # #1a1b26
    accent_blue = (122, 162, 247, 255)  # #7aa2f7
    accent_purple = (187, 154, 247, 255) # #bb9af7
    accent_green = (158, 206, 106, 255) # #9ece6a
    text_color = (169, 177, 214, 255)   # #a9b1d6
    white = (255, 255, 255, 255)
    
    # Center and radius
    cx, cy = size // 2, size // 2
    radius = size // 2 - 20
    
    # Draw outer circle background
    draw.ellipse([cx - radius, cy - radius, cx + radius, cy + radius], fill=bg_color)
    
    # Draw subtle border ring
    border_width = 4
    draw.ellipse([cx - radius, cy - radius, cx + radius, cy + radius], 
                 outline=accent_blue, width=border_width)
    
    # Draw inner decorative ring
    inner_radius = radius - 30
    draw.ellipse([cx - inner_radius, cy - inner_radius, cx + inner_radius, cy + inner_radius], 
                 outline=(*accent_blue[:3], 80), width=1)
    
    # Draw the Yggdrasil tree (stylized)
    tree_color = accent_green
    trunk_color = (*accent_blue[:3], 200)
    
    # Tree trunk
    trunk_width = 8
    trunk_bottom = cy + 120
    trunk_top = cy - 80
    draw.line([(cx, trunk_bottom), (cx, trunk_top)], fill=trunk_color, width=trunk_width)
    
    # Draw roots (3 main roots)
    root_color = (*accent_blue[:3], 150)
    for angle_offset in [-40, 0, 40]:
        root_len = 60
        end_x = cx + math.sin(math.radians(angle_offset)) * root_len
        end_y = trunk_bottom + math.cos(math.radians(angle_offset)) * root_len * 0.5
        draw.line([(cx, trunk_bottom), (end_x, end_y)], fill=root_color, width=4)
    
    # Draw branches (5 main branches - representing the 5 realms)
    branch_color = accent_purple
    branches = [
        (-60, -100, 80),   # left upper
        (-30, -90, 70),    # left mid
        (0, -110, 90),     # top
        (30, -90, 70),     # right mid
        (60, -100, 80),    # right upper
    ]
    
    for angle, start_offset, length in branches:
        start_y = trunk_top + abs(start_offset) // 3
        end_x = cx + math.sin(math.radians(angle)) * length
        end_y = start_y - math.cos(math.radians(angle)) * length * 0.7
        draw.line([(cx, start_y), (end_x, end_y)], fill=branch_color, width=4)
        
        # Add leaves/nodes at branch ends
        node_size = 8
        draw.ellipse([end_x - node_size, end_y - node_size, 
                      end_x + node_size, end_y + node_size], 
                     fill=accent_green, outline=white, width=2)
    
    # Add network connection lines (mesh effect)
    mesh_color = (*accent_blue[:3], 60)
    nodes = []
    for angle, start_offset, length in branches:
        start_y = trunk_top + abs(start_offset) // 3
        end_x = cx + math.sin(math.radians(angle)) * length
        end_y = start_y - math.cos(math.radians(angle)) * length * 0.7
        nodes.append((end_x, end_y))
    
    # Connect some nodes with dotted lines (mesh network effect)
    for i in range(len(nodes)):
        for j in range(i + 2, len(nodes)):
            if (i + j) % 2 == 0:
                x1, y1 = nodes[i]
                x2, y2 = nodes[j]
                # Draw dotted line
                steps = 10
                for k in range(0, steps, 2):
                    t1 = k / steps
                    t2 = (k + 1) / steps
                    px1 = x1 + (x2 - x1) * t1
                    py1 = y1 + (y2 - y1) * t1
                    px2 = x1 + (x2 - x1) * t2
                    py2 = y1 + (y2 - y1) * t2
                    draw.line([(px1, py1), (px2, py2)], fill=mesh_color, width=2)
    
    # Add lightning bolt symbol at the base
    bolt_color = (224, 175, 104, 255)  # #e0af68 (warning/amber)
    bolt_cx = cx
    bolt_cy = cy + 60
    bolt_size = 25
    
    # Lightning bolt shape
    bolt_points = [
        (bolt_cx - 8, bolt_cy - bolt_size),
        (bolt_cx + 12, bolt_cy - bolt_size),
        (bolt_cx + 2, bolt_cy - 5),
        (bolt_cx + 15, bolt_cy - 5),
        (bolt_cx - 5, bolt_cy + bolt_size),
        (bolt_cx + 5, bolt_cy + 5),
        (bolt_cx - 8, bolt_cy + 5),
    ]
    draw.polygon(bolt_points, fill=bolt_color)
    
    # Draw text - "YGGDRASIL" at top
    try:
        # Try to use a nice font, fall back to default
        font_large = ImageFont.truetype("arial.ttf", 42)
        font_small = ImageFont.truetype("arial.ttf", 28)
        font_tiny = ImageFont.truetype("arial.ttf", 20)
    except:
        font_large = ImageFont.load_default()
        font_small = ImageFont.load_default()
        font_tiny = ImageFont.load_default()
    
    # Top arc text - "YGGDRASIL"
    text = "YGGDRASIL"
    text_color_top = accent_blue
    
    # Calculate positions for arc text
    arc_radius = radius - 50
    total_angle = 120  # degrees
    start_angle = 270 - total_angle / 2
    
    for i, char in enumerate(text):
        angle = start_angle + (i / (len(text) - 1)) * total_angle
        angle_rad = math.radians(angle)
        x = cx + arc_radius * math.cos(angle_rad)
        y = cy + arc_radius * math.sin(angle_rad)
        
        # Rotate character
        char_img = Image.new('RGBA', (50, 50), (0, 0, 0, 0))
        char_draw = ImageDraw.Draw(char_img)
        bbox = char_draw.textbbox((0, 0), char, font=font_large)
        char_w = bbox[2] - bbox[0]
        char_h = bbox[3] - bbox[1]
        char_draw.text(((50 - char_w) // 2, (50 - char_h) // 2), char, fill=text_color_top, font=font_large)
        
        rotated = char_img.rotate(-(angle - 270), expand=True, resample=Image.BICUBIC)
        paste_x = int(x - rotated.width // 2)
        paste_y = int(y - rotated.height // 2)
        
        img.paste(rotated, (paste_x, paste_y), rotated)
    
    # Bottom arc text - "MESH CHAT"
    bottom_text = "MESH CHAT"
    text_color_bottom = accent_purple
    
    bottom_arc_radius = radius - 50
    bottom_total_angle = 100
    bottom_start_angle = 90 - bottom_total_angle / 2
    
    for i, char in enumerate(bottom_text):
        angle = bottom_start_angle + (i / (len(bottom_text) - 1)) * bottom_total_angle
        angle_rad = math.radians(angle)
        x = cx + bottom_arc_radius * math.cos(angle_rad)
        y = cy + bottom_arc_radius * math.sin(angle_rad)
        
        char_img = Image.new('RGBA', (50, 50), (0, 0, 0, 0))
        char_draw = ImageDraw.Draw(char_img)
        bbox = char_draw.textbbox((0, 0), char, font=font_small)
        char_w = bbox[2] - bbox[0]
        char_h = bbox[3] - bbox[1]
        char_draw.text(((50 - char_w) // 2, (50 - char_h) // 2), char, fill=text_color_bottom, font=font_small)
        
        rotated = char_img.rotate(-(angle - 270), expand=True, resample=Image.BICUBIC)
        paste_x = int(x - rotated.width // 2)
        paste_y = int(y - rotated.height // 2)
        
        img.paste(rotated, (paste_x, paste_y), rotated)
    
    # Add small decorative dots around the circle
    for angle in range(0, 360, 15):
        dot_radius = radius - 15
        angle_rad = math.radians(angle)
        x = cx + dot_radius * math.cos(angle_rad)
        y = cy + dot_radius * math.sin(angle_rad)
        dot_size = 3
        draw.ellipse([x - dot_size, y - dot_size, x + dot_size, y + dot_size], 
                     fill=(*accent_blue[:3], 100))
    
    return img


def main():
    """Generate and save the logo"""
    print("Generating Yggdrasil Mesh Chat logo...")
    
    # Generate high-res logo
    logo = create_logo(800)
    
    # Save as PNG (for transparency)
    logo.save("logo.png", "PNG")
    print("Saved: logo.png")
    
    # Save as JPG (white background for compatibility)
    jpg_img = Image.new('RGB', logo.size, (255, 255, 255))
    jpg_img.paste(logo, mask=logo.split()[3])  # Use alpha channel as mask
    jpg_img.save("logo.jpg", "JPEG", quality=95)
    print("Saved: logo.jpg")
    
    # Also create a smaller version for README (160x160)
    small_logo = logo.resize((160, 160), Image.LANCZOS)
    small_logo.save("logo_small.png", "PNG")
    print("Saved: logo_small.png")
    
    print("\nLogo generation complete!")


if __name__ == "__main__":
    main()
