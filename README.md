# Клонируем репозиторий
git clone git@github.com:RPogodin/digital2026.git

# Для Ubuntu
sudo apt-get update
sudo apt-get install -y git
sudo apt-get install -y openssh-client

# Для Mac
# Сначала устанавливаем пакетный менеджер Homebrew
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" 
# И далее устанавливаем гит
brew install git

# На маке ssh-keygen должна быть из коробки, но, если её нет, можно поставить: brew install openssh

ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519_github -C "sand@sandbox github"

cat >> ~/.ssh/config <<'EOF'
Host github.com
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_github
  IdentitiesOnly yes
EOF
chmod 600 ~/.ssh/config

cat ~/.ssh/id_ed25519_github.pub

git config --global user.email "ruspog@gmail.com"
  git config --global user.name "Ruslan Pogodin"

git config --global push.autoSetupRemote true

# Создать ветку локально
git checkout -b your_name/name
# Или
git switch -c your_name/name

# Перепрыгнуть в другую ветку
git checkout branch_name
# Или
git switch branch_name

# Забрать все обновления с удалённого репозитория, не меняя текущую ветку
git fetch --all --prune 

# Стянуть все изменения текущей ветки
git pull

# # 3) Обновить master до состояния origin/master (обычный вариант через fast-forward)
git pull --ff-only origin master

# Добавить изменения в индекс
git add file.txt     # добавить конкретный файл
git add .            # добавить всё в текущей папке и внутри

# Убрать изменения из индекса
git restore --staged file.txt

# Сохранить изменения локально
git commit -m "Fix bug in auth"

# Отправить изменения на удалённый репозиторий (вместе с веткой)
git push

# Флоу работы (перед каждой новой задачей)
git fetch --all --prune
git switch master
