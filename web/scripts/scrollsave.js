 // Save scroll position when submit
    document.addEventListener("submit", () => {
    sessionStorage.setItem("scrollY", window.scrollY);
});

    // Restore scroll position
    window.addEventListener("load", () => {
    const y = sessionStorage.getItem("scrollY");
    if (y !== null) {
    window.scrollTo(0, parseInt(y, 10));
    sessionStorage.removeItem("scrollY");
}
});
