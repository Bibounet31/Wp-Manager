// Wallpaper Modal with Comments
let currentWallpaperId = null;

function openWallpaperModal(wallpaperId, imageUrl, title, date) {
    currentWallpaperId = wallpaperId;

    const modal = document.getElementById('wallpaperModal');
    const modalImage = document.getElementById('modalImage');
    const modalTitle = document.getElementById('modalImageTitle');
    const modalDate = document.getElementById('modalImageDate');

    // Set image and info
    modalImage.src = imageUrl;
    modalImage.alt = title;
    modalTitle.textContent = title;
    modalDate.textContent = date;

    // Show modal
    modal.classList.add('active');
    document.body.style.overflow = 'hidden'; // Prevent background scrolling

    // Load comments for this wallpaper
    loadComments(wallpaperId);
}

function closeWallpaperModal() {
    const modal = document.getElementById('wallpaperModal');
    modal.classList.remove('active');
    document.body.style.overflow = ''; // Restore scrolling
    currentWallpaperId = null;

    // Clear comment form
    document.getElementById('commentText').value = '';
    updateCharCount();
}

// Close modal with Escape key
document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') {
        closeWallpaperModal();
    }
});

// Character counter for comment textarea
document.addEventListener('DOMContentLoaded', function() {
    const commentText = document.getElementById('commentText');
    if (commentText) {
        commentText.addEventListener('input', updateCharCount);
    }
});

function updateCharCount() {
    const commentText = document.getElementById('commentText');
    const charCount = document.getElementById('charCount');
    if (commentText && charCount) {
        const count = commentText.value.length;
        charCount.textContent = `${count}/500`;

        if (count > 450) {
            charCount.style.color = '#ff6b81';
        } else {
            charCount.style.color = '';
        }
    }
}

// Load comments from server
async function loadComments(wallpaperId) {
    const commentsList = document.getElementById('commentsList');

    try {
        const response = await fetch(`/api/comments/${wallpaperId}`);

        if (!response.ok) {
            throw new Error('Failed to load comments');
        }

        const comments = await response.json();

        if (comments && comments.length > 0) {
            commentsList.innerHTML = comments.map(comment => createCommentHTML(comment)).join('');
        } else {
            commentsList.innerHTML = `
                <div class="no-comments">
                    <p>No comments yet. Be the first to comment! ✨</p>
                </div>
            `;
        }
    } catch (error) {
        console.error('Error loading comments:', error);
        commentsList.innerHTML = `
            <div class="no-comments">
                <p>No comments yet. Be the first to comment! ✨</p>
            </div>
        `;
    }
}

// Create HTML for a single comment
function createCommentHTML(comment) {
    const date = new Date(comment.created_at);
    const timeAgo = getTimeAgo(date);

    return `
        <div class="comment-item" data-comment-id="${comment.id}">
            <div class="comment-header">
                <span class="comment-author">${escapeHtml(comment.username)}</span>
                <span class="comment-time">${timeAgo}</span>
            </div>
            <div class="comment-body">
                ${escapeHtml(comment.text)}
            </div>
        </div>
    `;
}

// Submit new comment
async function submitComment(event) {
    event.preventDefault();

    const commentText = document.getElementById('commentText');
    const text = commentText.value.trim();

    if (!text || !currentWallpaperId) {
        return;
    }

    try {
        const response = await fetch('/api/comments', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                wallpaper_id: currentWallpaperId,
                text: text
            })
        });

        if (!response.ok) {
            throw new Error('Failed to post comment');
        }

        // Clear form
        commentText.value = '';
        updateCharCount();

        // Reload comments
        loadComments(currentWallpaperId);

    } catch (error) {
        console.error('Error posting comment:', error);
        alert('Failed to post comment. Please try again.');
    }
}

// Helper: Get time ago string
function getTimeAgo(date) {
    const seconds = Math.floor((new Date() - date) / 1000);

    const intervals = {
        year: 31536000,
        month: 2592000,
        week: 604800,
        day: 86400,
        hour: 3600,
        minute: 60
    };

    for (const [unit, secondsInUnit] of Object.entries(intervals)) {
        const interval = Math.floor(seconds / secondsInUnit);
        if (interval >= 1) {
            return interval === 1 ? `1 ${unit} ago` : `${interval} ${unit}s ago`;
        }
    }

    return 'just now';
}

// Helper: Escape HTML to prevent XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}