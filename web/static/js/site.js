// Alpine.js components for personal website

document.addEventListener('alpine:init', () => {
    // Project filter on projects page
    Alpine.data('projectFilter', () => ({
        selectedStack: 'all',

        filterProjects() {
            // htmx will handle the request
            htmx.ajax('GET', `/projects?stack=${this.selectedStack}`, {
                target: '#project-grid',
                swap: 'innerHTML'
            });
        }
    }));

    // Blog tag filter
    Alpine.data('blogFilter', () => ({
        selectedTag: 'all',

        filterPosts() {
            htmx.ajax('GET', `/blog?tag=${this.selectedTag}`, {
                target: '#blog-list',
                swap: 'innerHTML'
            });
        }
    }));
});
