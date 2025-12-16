async function getUsername(userId) {
    try {
        const res = await fetch(`http://localhost:8080/api/v1/users/${userId}`);
        if (res.ok) {
            const data = await res.json();
            return data.data.username;
        } else {
            return `User ${userId}`;
        }
    } catch (e) {
        return `User ${userId}`;
    }
}

async function loadPosts() {
    const res = await fetch("http://localhost:8080/api/v1/posts");
    const data = await res.json();
    const output = document.getElementById("output");
    output.innerHTML = '';
    if (data.data && Array.isArray(data.data)) {
        for (const post of data.data) {
            const username = await getUsername(post.author_id);
            const postDiv = document.createElement('div');
            postDiv.className = 'post';
            postDiv.innerHTML = `
                <div class="header">
                    <div class="left">
                        <div class="author">Author: ${username}</div>
                        <div class="title">${post.title}</div>
                    </div>
                    <div class="right">
                        <div class="date">${post.created_at.slice(0, -3)}</div>
                    </div>
                </div>
                <div class="body">${post.body}</div>
                <div class="bottom">
                    <button class="load-comments" onclick="toggleComments(this, ${post.id})">Load comments</button>
                </div>
                <div class="comments" id="comments-${post.id}"></div>
            `;
            output.appendChild(postDiv);
        }
    }
}

async function toggleComments(button, postId) {
    const commentsDiv = document.getElementById(`comments-${postId}`);
    if (!commentsDiv.innerHTML.trim()) {
        // Load comments
        try {
            const res = await fetch(`http://localhost:8080/api/v1/posts/${postId}/comments`);
            const data = await res.json();
            commentsDiv.innerHTML = '';
            // Add comment input
            const inputDiv = document.createElement('div');
            inputDiv.innerHTML = `<textarea placeholder="Join the conversation" class="comment-input"></textarea>`;
            commentsDiv.appendChild(inputDiv);
            const textarea = inputDiv.querySelector('.comment-input');
            textarea.addEventListener('input', function() {
                this.style.height = 'auto';
                this.style.height = this.scrollHeight + 'px';
            });
            textarea.addEventListener('keyup', function(e) {
                if (e.key === 'Enter') {
                    this.style.height = 'auto';
                    this.style.height = this.scrollHeight + 'px';
                }
            });
            let hasComments = false;
            if (data.data && Array.isArray(data.data)) {
                for (const comment of data.data) {
                    const commentUsername = await getUsername(comment.user_id);
                    const commentDiv = document.createElement('div');
                    commentDiv.className = 'comment';
                    commentDiv.innerHTML = `<strong>${commentUsername}</strong>: ${comment.body}`;
                    commentsDiv.appendChild(commentDiv);
                    hasComments = true;
                }
            }
            if (hasComments) {
                button.textContent = 'Hide comments';
            }
        } catch (e) {
            console.error(e);
        }
    } else {
        // Hide comments
        commentsDiv.innerHTML = '';
        button.textContent = 'Load comments';
    }
}

