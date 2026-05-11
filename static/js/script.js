var uploadArea = document.getElementById('uploadArea');
var fileInput = document.getElementById('fileInput');
var uploadBtn = document.getElementById('uploadBtn');
var loading = document.getElementById('loading');
var error = document.getElementById('error');
var results = document.getElementById('results');
var resultsTitle = document.getElementById('resultsTitle');
var fileList = document.getElementById('fileList');
var actions = document.getElementById('actions');
var downloadAllBtn = document.getElementById('downloadAllBtn');
var uploadAnotherBtn = document.getElementById('uploadAnotherBtn');
var step1 = document.getElementById('step1');
var step2 = document.getElementById('step2');
var notices = document.querySelectorAll('#step2 .notice');

var selectedFile = null;
var sessionData = null;

function selectFile(file) {
    hideError();
    if (!file.name.toLowerCase().endsWith('.pdf')) {
        showError('Please upload a PDF file');
        return;
    }
    selectedFile = file;
    uploadBtn.disabled = false;
    uploadArea.querySelector('.upload-text').textContent = file.name;
}

function handleUpload(file) {
    hideError();
    results.classList.remove('show');
    actions.classList.remove('show');
    downloadAllBtn.style.display = 'none';
    step1.classList.remove('active');
    step2.classList.add('active');
    loading.classList.add('show');

    var formData = new FormData();
    formData.append('pdf', file);

    fetch('/upload', {
        method: 'POST',
        body: formData
    })
    .then(function(response) {
        return response.json();
    })
    .then(function(data) {
        loading.classList.remove('show');
        if (data.error) {
            showError(data.error);
            actions.classList.add('show');
            return;
        }
        sessionData = data;
        showResults(data);
    })
    .catch(function() {
        loading.classList.remove('show');
        showError('Failed to process PDF');
        actions.classList.add('show');
    });
}

function showResults(data) {
    results.classList.add('show');
    fileList.innerHTML = '';

    if (data.files.length === 0) {
        resultsTitle.textContent = 'No embedded files found in this PDF.';
        downloadAllBtn.style.display = 'none';
        actions.classList.add('show');
        return;
    }

    resultsTitle.textContent = `Found ${data.files.length} embedded file${data.files.length > 1 ? 's' : ''}.`;

    // Show/hide notices based on results
    notices.forEach(function(n) {
        n.style.display = data.files.length > 0 ? 'block' : 'none';
    });

    data.files.forEach(function(filename) {
        var li = document.createElement('li');
        var span = document.createElement('span');
        span.textContent = filename;

        var btn = document.createElement('button');
        btn.className = 'download-btn';
        btn.textContent = 'Download';
        btn.onclick = () => {
            window.location.href = `/download?id=${data.id}&filename=${encodeURIComponent(filename)}`;
        };

        li.appendChild(span);
        li.appendChild(btn);
        fileList.appendChild(li);
    });

    if (data.hasZip) {
        downloadAllBtn.style.display = 'block';
    } else {
        downloadAllBtn.style.display = 'none';
    }
    actions.classList.add('show');
}

function showError(msg) {
    error.textContent = msg;
    error.classList.add('show');
}

function hideError() {
    error.classList.remove('show');
}

function resetToStep1() {
    step2.classList.remove('active');
    step1.classList.add('active');
    selectedFile = null;
    sessionData = null;
    uploadBtn.disabled = true;
    fileInput.value = '';
    results.classList.remove('show');
    actions.classList.remove('show');
    hideError();
    notices.forEach(function(n) {
        n.style.display = 'none';
    });
    uploadArea.querySelector('.upload-text').textContent = 'Drop your PDF here or click to browse';
}

uploadArea.addEventListener('click', function() {
    fileInput.click();
});

uploadArea.addEventListener('dragover', function(e) {
    e.preventDefault();
    uploadArea.classList.add('dragover');
});

uploadArea.addEventListener('dragleave', function() {
    uploadArea.classList.remove('dragover');
});

uploadArea.addEventListener('drop', function(e) {
    e.preventDefault();
    uploadArea.classList.remove('dragover');
    var file = e.dataTransfer.files[0];
    if (file) selectFile(file);
});

fileInput.addEventListener('change', function(e) {
    var file = e.target.files[0];
    if (file) selectFile(file);
});

uploadBtn.addEventListener('click', function() {
    if (selectedFile) handleUpload(selectedFile);
});

uploadAnotherBtn.addEventListener('click', resetToStep1);

downloadAllBtn.addEventListener('click', function() {
    if (!sessionData || !sessionData.id || !sessionData.hasZip) return;
    window.location.href = `/download?id=${sessionData.id}&all=true`;
});
