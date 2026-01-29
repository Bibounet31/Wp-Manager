# Wallpaper Manager

Wallpaper management application built with Go and a MySQL database.  
Each user can manage their own wallpapers and interact with others through social features.

## Technologies
- Go
- MySQL

## Features

### Users
- Account creation
- Login
- User profile (profile picture, bio)
- Private wallpapers per user
- Ability to make a wallpaper public after admin approval

### Wallpaper Management
- Upload
- Delete
- Archive
- Focus on personalization

### Organization
- Search
- Sort by theme or tags
- Create collections
- Favorites system
- Dominant color detection
- Phone and PC format support (resolution)

### Social
- Rating system (e.g. “wallpaper of the month”)
- Comments
- Friend request system to share wallpapers
- Notifications
- User badges

## TODO
- [V] Go project initialization
- [V] MySQL database connection
- [~] Authentication system
- [~] Wallpaper upload
- [ ] Private/Public management with admin approval
- [ ] Tags and collections
- [ ] Search and sorting
- [~] Favorites
- [ ] Ratings and comments
- [ ] Friend system
- [ ] Notifications
- [~] User profiles
- [ ] Badges
- [~] Rename WAllpapers
- [V] since pictures aren't saved when the container reboots, we need to clear the db where the pics are missing on start. (to not have only the name with a missing picture)
## Goal
Build a customizable wallpaper management and sharing platform with an exhibition-style layout and community features.
