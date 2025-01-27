git add --all && \
git commit -m "$1" && \
git push origin $2 && \
git checkout main && \
git merge $2 && \
git push origin main && \
git checkout $2 && \
git tag $3 && \
git push origin $3