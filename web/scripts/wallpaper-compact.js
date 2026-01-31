// Auto-apply compact mode for short images
document.addEventListener('DOMContentLoaded', function() {
    function checkImageHeights() {
        const wallpaperCards = document.querySelectorAll('.wallpaper-card');

        wallpaperCards.forEach(card => {
            const image = card.querySelector('.wallpaper-image');

            if (image) {
                // Wait for image to load to get accurate height
                if (image.complete) {
                    applyCompactMode(card, image);
                } else {
                    image.addEventListener('load', function() {
                        applyCompactMode(card, image);
                    });
                }
            }
        });
    }

    function applyCompactMode(card, image) {
        const imageHeight = image.offsetHeight;

        // If image is less than 280px tall, use compact mode (emoji only)
        if (imageHeight < 280) {
            card.classList.add('compact-mode');
        } else {
            card.classList.remove('compact-mode');
        }
    }

    // Initial check
    checkImageHeights();

    // Recheck on window resize
    let resizeTimer;
    window.addEventListener('resize', function() {
        clearTimeout(resizeTimer);
        resizeTimer = setTimeout(checkImageHeights, 250);
    });
});