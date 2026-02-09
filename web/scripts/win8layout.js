/**
 * Win8 Masonry Layout - FINAL FIX
 * Calculates grid spans based on IMAGE CONTAINER height (not card height)
 * since the overlay is position:absolute and doesn't affect card height
 */
(function() {
    console.log('Win8Layout FINAL: Script starting...');

    function initWin8Layout() {
        const grid = document.querySelector('.spell-grid');
        if (!grid) {
            console.log('Win8Layout: Grid not found');
            return;
        }

        const cards = grid.querySelectorAll('.wallpaper-card');
        if (cards.length === 0) {
            console.log('Win8Layout: No cards found');
            return;
        }

        console.log(`Win8Layout: Found ${cards.length} cards`);

        function resizeGridItems() {
            const gridStyles = getComputedStyle(grid);
            const gap = parseInt(gridStyles.gap) || 40;
            const rowHeight = parseInt(gridStyles.gridAutoRows) || 10;

            console.log(`Win8Layout: Gap=${gap}px, RowHeight=${rowHeight}px`);

            cards.forEach((card, index) => {
                const imageContainer = card.querySelector('.wallpaper-image-container');
                if (!imageContainer) {
                    console.log(`Card ${index}: No image container found`);
                    return;
                }

                // Get the HEIGHT OF THE IMAGE CONTAINER (not the card!)
                const containerHeight = imageContainer.getBoundingClientRect().height;

                // Calculate row span based on image container height
                const rowSpan = Math.ceil((containerHeight + gap) / (rowHeight + gap));

                card.style.gridRowEnd = `span ${rowSpan}`;

                console.log(`Card ${index}: containerHeight=${containerHeight}px, rowSpan=${rowSpan}`);
            });

            console.log('Win8Layout: Layout recalculated');
        }

        // Wait for all images to load
        const images = Array.from(grid.querySelectorAll('.wallpaper-image'));
        console.log(`Win8Layout: Waiting for ${images.length} images to load`);

        let loadedCount = 0;
        const totalImages = images.length;

        function onImageLoad() {
            loadedCount++;
            console.log(`Win8Layout: Image ${loadedCount}/${totalImages} loaded`);

            if (loadedCount === totalImages) {
                console.log('Win8Layout: All images loaded, calculating layout...');
                // Multiple recalculations
                setTimeout(resizeGridItems, 100);
                setTimeout(resizeGridItems, 300);
                setTimeout(resizeGridItems, 600);
                setTimeout(resizeGridItems, 1000);
            }
        }

        if (totalImages === 0) {
            console.log('Win8Layout: No images');
            resizeGridItems();
        } else {
            images.forEach((img, idx) => {
                if (img.complete && img.naturalHeight > 0) {
                    console.log(`Image ${idx} already loaded (${img.naturalWidth}x${img.naturalHeight})`);
                    onImageLoad();
                } else {
                    img.addEventListener('load', function() {
                        console.log(`Image ${idx} loaded (${this.naturalWidth}x${this.naturalHeight})`);
                        onImageLoad();
                    });
                    img.addEventListener('error', () => {
                        console.log(`Image ${idx} failed to load`);
                        onImageLoad();
                    });
                }
            });
        }

        // Recalculate on window resize
        let resizeTimeout;
        window.addEventListener('resize', () => {
            clearTimeout(resizeTimeout);
            resizeTimeout = setTimeout(() => {
                console.log('Win8Layout: Window resized, recalculating...');
                resizeGridItems();
            }, 200);
        });
    }

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initWin8Layout);
    } else {
        initWin8Layout();
    }
})();