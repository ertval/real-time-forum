const API_BASE = 'http://localhost:8080/api/v1';

document.addEventListener('DOMContentLoaded', function() {
    if (document.getElementById('posts-list')) {
        loadPosts();
    } else if (document.getElementById('post-details')) {
        loadPost();
    }
});

async function loadPosts(page = 1) {
    try {
        const response = await fetch(`${API_BASE}/posts?page=${page}&per_page=10`);
        const data = await response.json();
        if (data.error) {
            console.error(data.error);
            return;
        }
        const postsList = document.getElementById('posts-list');
        postsList.innerHTML = '';
        data.data.forEach(post => {
            const postDiv = document.createElement('div');
            postDiv.className = 'post';
            postDiv.innerHTML = `
                <h3><a href="/posts/${post.id}">${post.title}</a></h3>
                <p>${post.body}</p>
                <small>By User ${post.author_id} on ${new Date(post.created_at).toLocaleString()}</small>
            `;
            postsList.appendChild(postDiv);
        });
        // Pagination
        const meta = data.meta;
        const pagination = document.getElementById('pagination');
        pagination.innerHTML = '';
        if (meta.page > 1) {
            const prev = document.createElement('button');
            prev.textContent = 'Previous';
            prev.onclick = () => loadPosts(meta.page - 1);
            pagination.appendChild(prev);
        }
        if (meta.page * meta.per_page < meta.total) {
            const next = document.createElement('button');
            next.textContent = 'Next';
            next.onclick = () => loadPosts(meta.page + 1);
            pagination.appendChild(next);
        }
    } catch (error) {
        console.error('Error loading posts:', error);
    }
}

async function loadPost() {
    const pathParts = window.location.pathname.split('/');
    const postId = pathParts[pathParts.length - 1];
    if (!postId || isNaN(postId)) {
        document.getElementById('post-details').innerHTML = '<p>Invalid post ID</p>';
        return;
    }
    try {
        // Load post
        const postResponse = await fetch(`${API_BASE}/posts/${postId}`);
        const postData = await postResponse.json();
        if (postData.error) {
            document.getElementById('post-details').innerHTML = '<p>Post not found</p>';
            return;
        }
        const post = postData.data;
        document.getElementById('post-details').innerHTML = `
            <h2>${post.title}</h2>
            <p>${post.body}</p>
            <small>By User ${post.author_id} on ${new Date(post.created_at).toLocaleString()}</small>
        `;

        // Load comments
        const commentsResponse = await fetch(`${API_BASE}/posts/${postId}/comments?page=1&per_page=20`);
        const commentsData = await commentsResponse.json();
        const commentsList = document.getElementById('comments-list');
        commentsList.innerHTML = '';
        if (commentsData.data && commentsData.data.length > 0) {
            commentsData.data.forEach(comment => {
                const commentDiv = document.createElement('div');
                commentDiv.className = 'comment';
                commentDiv.innerHTML = `
                    <p>${comment.body}</p>
                    <small>By User ${comment.user_id} on ${new Date(comment.created_at).toLocaleString()}</small>
                `;
                commentsList.appendChild(commentDiv);
            });
        } else {
            commentsList.innerHTML = '<p>No comments yet.</p>';
        }
    } catch (error) {
        console.error('Error loading post:', error);
    }
}