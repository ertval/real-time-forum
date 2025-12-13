async function loadPosts() {
    const res = await fetch("http://localhost:8080/api/v1/posts");
    const data = await res.json();
    const output = document.getElementById("output");
    output.innerHTML = '';
    if (data.data && Array.isArray(data.data)) {
        data.data.forEach(post => {
            const postDiv = document.createElement('div');
            postDiv.className = 'post';
            postDiv.innerHTML = `
                <div class="header">
                    <div class="left">
                        <div class="author">Author: ${post.author_id}</div>
                        <div class="title">${post.title}</div>
                    </div>
                    <div class="right">
                        <div class="date">${post.created_at.slice(0, -3)}</div>
                    </div>
                </div>
                <div class="body">${post.body}</div>
                <div class="bottom">
                    <button class="load-comments" onclick="loadComments(${post.id})">Load comments</button>
                </div>
                <div class="comments" id="comments-${post.id}"></div>
            `;
            output.appendChild(postDiv);
        });
    }
}

function loadComments(postId) {
    fetch(`http://localhost:8080/api/v1/posts/${postId}/comments`)
        .then(res => res.json())
        .then(data => {
            const commentsDiv = document.getElementById(`comments-${postId}`);
            commentsDiv.innerHTML = '';
            if (data.data && Array.isArray(data.data)) {
                data.data.forEach(comment => {
                    const commentDiv = document.createElement('div');
                    commentDiv.className = 'comment';
                    commentDiv.innerHTML = `<strong>${comment.author_id}</strong>: ${comment.body}`;
                    commentsDiv.appendChild(commentDiv);
                });
            }
        });
}

