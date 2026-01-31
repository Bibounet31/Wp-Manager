// Rename wallpaper functionality
document.addEventListener('DOMContentLoaded', function() {
    const renameButtons = document.querySelectorAll('.rename-button');

    renameButtons.forEach(button => {
        button.addEventListener('click', function(e) {
            e.preventDefault();

            const wallpaperId = this.dataset.wallpaperId;
            const card = this.closest('.wallpaper-card');
            const currentName = card.querySelector('h3').textContent;

            // Show a prompt for new name
            const newName = prompt('Enter new name for wallpaper:', currentName);

            if (newName && newName.trim() !== '') {
                // Create and submit a form
                const form = document.createElement('form');
                form.method = 'POST';
                form.action = '/rename';

                // Add wallpaper ID
                const idInput = document.createElement('input');
                idInput.type = 'hidden';
                idInput.name = 'wallpaper_id';
                idInput.value = wallpaperId;

                // Add new name
                const nameInput = document.createElement('input');
                nameInput.type = 'hidden';
                nameInput.name = 'new_name';
                nameInput.value = newName.trim();

                form.appendChild(idInput);
                form.appendChild(nameInput);
                document.body.appendChild(form);
                form.submit();
            }
        });
    });
});