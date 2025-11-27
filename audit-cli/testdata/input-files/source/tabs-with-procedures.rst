============================
Tabs Containing Procedures
============================

This file tests tabs that contain different procedures.

Installation Instructions
==========================

Choose your operating system:

.. tabs::

   .. tab:: macOS
      :tabid: macos

      .. procedure::

         .. step:: Install Homebrew

            If you don't have Homebrew installed:

            .. code-block:: bash

               /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

         .. step:: Install MongoDB

            Use Homebrew to install MongoDB:

            .. code-block:: bash

               brew tap mongodb/brew
               brew install mongodb-community

         .. step:: Start MongoDB

            Start the MongoDB service:

            .. code-block:: bash

               brew services start mongodb-community

   .. tab:: Ubuntu
      :tabid: ubuntu

      .. procedure::

         .. step:: Import the public key

            Import the MongoDB public GPG key:

            .. code-block:: bash

               curl -fsSL https://www.mongodb.org/static/pgp/server-8.0.asc | \
                  sudo gpg -o /usr/share/keyrings/mongodb-server-8.0.gpg \
                  --dearmor

         .. step:: Create the list file

            Create the list file for Ubuntu:

            .. code-block:: bash

               echo "deb [ signed-by=/usr/share/keyrings/mongodb-server-8.0.gpg ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/8.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-8.0.list

         .. step:: Install MongoDB

            Update the package database and install:

            .. code-block:: bash

               sudo apt-get update
               sudo apt-get install -y mongodb-org

         .. step:: Start MongoDB

            Start the MongoDB service:

            .. code-block:: bash

               sudo systemctl start mongod

   .. tab:: Windows
      :tabid: windows

      .. procedure::

         .. step:: Download the installer

            Download the MongoDB MSI installer from the MongoDB Download Center.

         .. step:: Run the installer

            Double-click the downloaded `.msi` file and follow the installation wizard.

         .. step:: Start MongoDB

            Start MongoDB as a Windows service:

            .. code-block:: powershell

               net start MongoDB

